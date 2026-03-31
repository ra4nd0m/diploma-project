using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Text;
using AuthService.Models;
using Microsoft.AspNetCore.Identity;
using Microsoft.IdentityModel.Tokens;

namespace AuthService.Services;

public class JwtService(
    ILogger<JwtService> logger,
    IConfiguration configuration,
    UserManager<User> userManager) : IJwtService
{
    public async Task<string> GenerateJwtToken(User user)
    {
        logger.LogInformation("Generating JWT token for user {email}", ObfuscateEmail(user.Email));

        var key = configuration["Jwt:Key"]
            ?? throw new InvalidOperationException("JWT key is not configured");

        var tokenExpiryRaw = configuration["Jwt:TokenExpiryMinutes"]
            ?? throw new InvalidOperationException("Token expiration is not configured");

        if (!int.TryParse(tokenExpiryRaw, out var tokenExpiryMinutes))
        {
            throw new InvalidOperationException("Token expiration is invalid");
        }

        var secretKey = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(key));
        var signingCredentials = new SigningCredentials(secretKey, SecurityAlgorithms.HmacSha256);

        var roles = await userManager.GetRolesAsync(user);
        var isStudent = roles.Any(role => string.Equals(role, "Student", StringComparison.OrdinalIgnoreCase));
        var isTeacher = roles.Any(role => string.Equals(role, "Teacher", StringComparison.OrdinalIgnoreCase));

        var email = user.Email ?? throw new InvalidOperationException("User email is not set");

        var claims = new List<Claim>
        {
            new(ClaimTypes.Name, email),
            new(ClaimTypes.NameIdentifier, user.Id),
            new("IsStudent", isStudent.ToString()),
            new("IsTeacher", isTeacher.ToString())
        };

        foreach (var role in roles)
        {
            claims.Add(new Claim(ClaimTypes.Role, role));
        }

        var token = new JwtSecurityToken(
            issuer: configuration["Jwt:Issuer"],
            audience: configuration["Jwt:Audience"],
            claims: claims,
            expires: DateTime.UtcNow.AddMinutes(tokenExpiryMinutes),
            signingCredentials: signingCredentials
        );

        return new JwtSecurityTokenHandler().WriteToken(token);
    }

    private static string ObfuscateEmail(string? email)
    {
        if (string.IsNullOrWhiteSpace(email))
        {
            return "<empty>";
        }

        var atIndex = email.IndexOf('@');
        if (atIndex <= 1)
        {
            return "***";
        }

        return $"{email[0]}***{email[(atIndex - 1)..]}";
    }
}
