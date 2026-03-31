using AuthService.Dtos;
using AuthService.Data;
using AuthService.Models;
using AuthService.Services;
using AuthService.Utils;
using Microsoft.AspNetCore.Identity;
using Microsoft.EntityFrameworkCore;

namespace AuthService.Routes;

public static class AuthRoutes
{
    public static void MapAuthRoutes(this WebApplication app)
    {
        var loggerFactory = app.Services.GetRequiredService<ILoggerFactory>();
        var logger = loggerFactory.CreateLogger("AuthRoutes");

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

                return Results.Ok(new { Token = accessToken });
            }
            catch (Exception ex)
            {
                logger.LogError(ex, "Failed to login user with email {Email}", EmailObfuscator.ObfuscateEmail(loginDto.Email));
                return Results.StatusCode(StatusCodes.Status500InternalServerError);
            }
        });

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

                return Results.Ok(new { Token = newAccessToken });
            }
            catch (Exception ex)
            {
                logger.LogError(ex, "Failed to refresh token");
                return Results.StatusCode(StatusCodes.Status500InternalServerError);
            }
        });

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
        });
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
}
