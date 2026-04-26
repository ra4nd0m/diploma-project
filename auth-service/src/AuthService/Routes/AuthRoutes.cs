using AuthService.Dtos;
using AuthService.Data;
using AuthService.Models;
using AuthService.Services;
using AuthService.Utils;
using System.Security.Claims;
using Microsoft.AspNetCore.Identity;
using Microsoft.EntityFrameworkCore;

namespace AuthService.Routes;

public static class AuthRoutes
{
    public static void MapAuthRoutes(this WebApplication app)
    {
        var loggerFactory = app.Services.GetRequiredService<ILoggerFactory>();
        var logger = loggerFactory.CreateLogger("AuthRoutes");

        app.MapPost("/register", async (
            RegisterDto registerDto,
            UserManager<User> userManager,
            RoleManager<IdentityRole> roleManager,
            IJwtService jwtService,
            IRefreshTokenService refreshTokenService,
            AuthDbContext dbContext,
            HttpContext context) =>
        {
            try
            {
                logger.LogInformation("Registration attempt for email {Email}", EmailObfuscator.ObfuscateEmail(registerDto.Email));

                var roleName = NormalizeSupportedRole(registerDto.Role);
                if (roleName is null)
                {
                    return Results.BadRequest("Role must be Student or Teacher");
                }

                var existingUser = await userManager.FindByEmailAsync(registerDto.Email);
                if (existingUser is not null)
                {
                    return Results.BadRequest("Email is already registered");
                }

                var appRole = await GetOrCreateAppRoleAsync(dbContext, roleName);
                await EnsureIdentityRoleExistsAsync(roleManager, roleName);

                var user = new User
                {
                    UserName = registerDto.Email,
                    Email = registerDto.Email,
                    DisplayName = registerDto.DisplayName,
                    SchoolName = registerDto.SchoolName,
                    RoleId = appRole.Id
                };

                var createResult = await userManager.CreateAsync(user, registerDto.Password);
                if (!createResult.Succeeded)
                {
                    var errors = createResult.Errors.Select(e => e.Description).ToArray();
                    return Results.BadRequest(new { Errors = errors });
                }

                var roleResult = await userManager.AddToRoleAsync(user, roleName);
                if (!roleResult.Succeeded)
                {
                    var errors = roleResult.Errors.Select(e => e.Description).ToArray();
                    await userManager.DeleteAsync(user);
                    return Results.BadRequest(new { Errors = errors });
                }

                logger.LogInformation("User {Email} registered with role {Role}", EmailObfuscator.ObfuscateEmail(registerDto.Email), roleName);

                var accessToken = await jwtService.GenerateJwtToken(user);
                var refreshToken = await refreshTokenService.GenerateRefreshToken(user);

                context.Response.Cookies.Append("refreshToken", refreshToken.Token, BuildCookieOptions(refreshToken.Expires));

                return Results.Ok(new AuthTokenResponse(accessToken));
            }
            catch (Exception ex)
            {
                logger.LogError(ex, "Failed to register user with email {Email}", EmailObfuscator.ObfuscateEmail(registerDto.Email));
                return Results.StatusCode(StatusCodes.Status500InternalServerError);
            }
        })
        .Accepts<RegisterDto>("application/json")
        .Produces<AuthTokenResponse>(StatusCodes.Status200OK, "application/json")
        .Produces<string>(StatusCodes.Status400BadRequest)
        .Produces(StatusCodes.Status500InternalServerError)
        .WithName("RegisterUser")
        .WithSummary("Register a new account")
        .WithDescription("Creates a new user account for either a student or teacher, assigns the requested role, and returns an access token while setting the refresh token cookie.")
        .WithTags("Auth");

        app.MapPost("/login", async (
            IJwtService jwtService,
            IRefreshTokenService refreshTokenService,
            UserManager<User> userManager,
            LoginDto loginDto,
            HttpContext context,
            AuthDbContext dbContext) =>
        {
            try
            {
                logger.LogInformation("Logging in user with email {Email}", EmailObfuscator.ObfuscateEmail(loginDto.Email));

                var user = await userManager.FindByEmailAsync(loginDto.Email);
                if (user is null || !await userManager.CheckPasswordAsync(user, loginDto.Password))
                {
                    logger.LogWarning("Failed to login user with email {Email}", EmailObfuscator.ObfuscateEmail(loginDto.Email));
                    return Results.BadRequest("Invalid email or password");
                }

                user = await dbContext.Users.FirstOrDefaultAsync(u => u.Id == user.Id);
                if (user is null)
                {
                    return Results.BadRequest("Invalid email or password");
                }

                logger.LogInformation("User {Email} logged in", EmailObfuscator.ObfuscateEmail(loginDto.Email));

                var accessToken = await jwtService.GenerateJwtToken(user);
                var refreshToken = await refreshTokenService.GenerateRefreshToken(user);

                context.Response.Cookies.Append("refreshToken", refreshToken.Token, BuildCookieOptions(refreshToken.Expires));

                return Results.Ok(new AuthTokenResponse(accessToken));
            }
            catch (Exception ex)
            {
                logger.LogError(ex, "Failed to login user with email {Email}", EmailObfuscator.ObfuscateEmail(loginDto.Email));
                return Results.StatusCode(StatusCodes.Status500InternalServerError);
            }
        })
        .Accepts<LoginDto>("application/json")
        .Produces<AuthTokenResponse>(StatusCodes.Status200OK, "application/json")
        .Produces<string>(StatusCodes.Status400BadRequest)
        .Produces(StatusCodes.Status500InternalServerError)
        .WithName("LoginUser")
        .WithSummary("Log in an existing user")
        .WithDescription("Validates the supplied credentials, issues a new access token, and refreshes the refresh token cookie.")
        .WithTags("Auth");

        app.MapPost("/refresh", async (
            IJwtService jwtService,
            IRefreshTokenService refreshTokenService,
            UserManager<User> userManager,
            HttpContext context,
            AuthDbContext dbContext) =>
        {
            try
            {
                logger.LogInformation("Token refresh attempt");

                if (!context.Request.Cookies.TryGetValue("refreshToken", out var currentToken) || string.IsNullOrWhiteSpace(currentToken))
                {
                    return Results.Unauthorized();
                }

                var rotatedRefreshToken = await refreshTokenService.RotateRefreshToken(currentToken);
                if (rotatedRefreshToken is null)
                {
                    logger.LogWarning("Invalid or reused refresh token used");
                    context.Response.Cookies.Delete("refreshToken");
                    return Results.Unauthorized();
                }

                var user = await userManager.FindByIdAsync(rotatedRefreshToken.UserId);
                if (user is null)
                {
                    context.Response.Cookies.Delete("refreshToken");
                    return Results.Unauthorized();
                }

                user = await dbContext.Users.FirstOrDefaultAsync(u => u.Id == user.Id);
                if (user is null)
                {
                    context.Response.Cookies.Delete("refreshToken");
                    return Results.Unauthorized();
                }

                var newAccessToken = await jwtService.GenerateJwtToken(user);
                context.Response.Cookies.Append("refreshToken", rotatedRefreshToken.Token, BuildCookieOptions(rotatedRefreshToken.Expires));

                return Results.Ok(new AuthTokenResponse(newAccessToken));
            }
            catch (Exception ex)
            {
                logger.LogError(ex, "Failed to refresh token");
                return Results.StatusCode(StatusCodes.Status500InternalServerError);
            }
        })
        .Produces<AuthTokenResponse>(StatusCodes.Status200OK, "application/json")
        .Produces(StatusCodes.Status401Unauthorized)
        .Produces(StatusCodes.Status500InternalServerError)
        .WithName("RefreshToken")
        .WithSummary("Refresh access token")
        .WithDescription("Uses the refresh token cookie to rotate the refresh token and issue a new access token.")
        .WithTags("Auth");

        app.MapPost("/logout", async (IRefreshTokenService refreshTokenService, HttpContext context) =>
        {
            try
            {
                logger.LogInformation("Logging out user");

                if (context.Request.Cookies.TryGetValue("refreshToken", out var token) && !string.IsNullOrWhiteSpace(token))
                {
                    await refreshTokenService.RecallRefreshToken(token);
                    context.Response.Cookies.Delete("refreshToken");
                }

                return Results.Ok();
            }
            catch (Exception ex)
            {
                logger.LogError(ex, "Failed to logout user");
                return Results.StatusCode(StatusCodes.Status500InternalServerError);
            }
        })
        .Produces(StatusCodes.Status200OK)
        .Produces(StatusCodes.Status500InternalServerError)
        .WithName("LogoutUser")
        .WithSummary("Log out the current user")
        .WithDescription("Revokes the refresh token cookie if present and ends the current session.")
        .WithTags("Auth");

        app.MapGet("/me", async (ClaimsPrincipal principal, AuthDbContext dbContext) =>
        {
            try
            {
                var userId = principal.FindFirstValue("sub");
                if (string.IsNullOrWhiteSpace(userId))
                {
                    return Results.Unauthorized();
                }

                var user = await dbContext.Users
                    .Include(u => u.Role)
                    .FirstOrDefaultAsync(u => u.Id == userId);

                if (user is null || string.IsNullOrWhiteSpace(user.Email))
                {
                    return Results.Unauthorized();
                }

                return Results.Ok(new MeResponseDto(
                    user.Id,
                    user.Email,
                    user.DisplayName,
                    user.SchoolName,
                    user.Role.Name));
            }
            catch (Exception ex)
            {
                logger.LogError(ex, "Failed to fetch profile info for current user");
                return Results.StatusCode(StatusCodes.Status500InternalServerError);
            }
        })
        .RequireAuthorization()
        .Produces<MeResponseDto>(StatusCodes.Status200OK, "application/json")
        .Produces(StatusCodes.Status401Unauthorized)
        .Produces(StatusCodes.Status500InternalServerError)
        .WithName("GetCurrentUser")
        .WithSummary("Get current user profile")
        .WithDescription("Returns profile and role information for the authenticated user.")
        .WithTags("Auth");
    }

    private static CookieOptions BuildCookieOptions(DateTime expires)
    {
        return new CookieOptions
        {
            HttpOnly = true,
            SameSite = SameSiteMode.Strict,
            Secure = true,
            Expires = expires
        };
    }

    private static string? NormalizeSupportedRole(string role)
    {
        if (string.Equals(role, "Student", StringComparison.OrdinalIgnoreCase))
        {
            return "Student";
        }

        if (string.Equals(role, "Teacher", StringComparison.OrdinalIgnoreCase))
        {
            return "Teacher";
        }

        return null;
    }

    private static async Task<Role> GetOrCreateAppRoleAsync(AuthDbContext dbContext, string roleName)
    {
        var code = roleName.ToUpperInvariant();

        var existingRole = await dbContext.Set<Role>()
            .FirstOrDefaultAsync(r => r.Code == code);

        if (existingRole is not null)
        {
            return existingRole;
        }

        var role = new Role
        {
            Name = roleName,
            Code = code
        };

        await dbContext.Set<Role>().AddAsync(role);
        await dbContext.SaveChangesAsync();

        return role;
    }

    private static async Task EnsureIdentityRoleExistsAsync(RoleManager<IdentityRole> roleManager, string roleName)
    {
        if (await roleManager.RoleExistsAsync(roleName))
        {
            return;
        }

        var createRoleResult = await roleManager.CreateAsync(new IdentityRole(roleName));
        if (!createRoleResult.Succeeded)
        {
            var errors = string.Join(", ", createRoleResult.Errors.Select(e => e.Description));
            throw new InvalidOperationException($"Could not create role {roleName}: {errors}");
        }
    }
}
