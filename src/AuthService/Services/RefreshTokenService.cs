using System.Security.Cryptography;
using AuthService.Data;
using AuthService.Models;
using Microsoft.EntityFrameworkCore;

namespace AuthService.Services;

public class RefreshTokenService(
    AuthDbContext context,
    IConfiguration configuration,
    ILogger<RefreshTokenService> logger) : IRefreshTokenService
{
    public async Task<RefreshToken> GenerateRefreshToken(User user)
    {
        var tokenExpiryDays = GetRefreshTokenExpiryDays();

        var token = new RefreshToken
        {
            UserId = user.Id,
            User = user,
            Token = CreateSecureToken(),
            Expires = DateTime.UtcNow.AddDays(tokenExpiryDays),
            CreatedAt = DateTime.UtcNow
        };

        await context.RefreshTokens.AddAsync(token);
        await context.SaveChangesAsync();

        return token;
    }

    public async Task<RefreshToken?> ValidateRefreshToken(string token)
    {
        var refreshToken = await context.RefreshTokens
            .Include(t => t.User)
            .FirstOrDefaultAsync(t => t.Token == token);

        if (refreshToken is null || !refreshToken.IsActive)
        {
            return null;
        }

        return refreshToken;
    }

    public async Task<RefreshToken?> RotateRefreshToken(string token)
    {
        var existingToken = await context.RefreshTokens
            .Include(t => t.User)
            .FirstOrDefaultAsync(t => t.Token == token);

        if (existingToken is null)
        {
            return null;
        }

        if (!existingToken.IsActive)
        {
            await RevokeTokenFamily(existingToken);
            return null;
        }

        var replacementToken = new RefreshToken
        {
            UserId = existingToken.UserId,
            User = existingToken.User,
            Token = CreateSecureToken(),
            Expires = DateTime.UtcNow.AddDays(GetRefreshTokenExpiryDays()),
            CreatedAt = DateTime.UtcNow
        };

        existingToken.RevokedAt = DateTime.UtcNow;
        existingToken.ReplacedByToken = replacementToken.Token;

        await context.RefreshTokens.AddAsync(replacementToken);
        await context.SaveChangesAsync();

        return replacementToken;
    }

    public async Task RecallRefreshToken(string token)
    {
        var refreshToken = await context.RefreshTokens.FirstOrDefaultAsync(t => t.Token == token);
        if (refreshToken is null || refreshToken.IsRevoked)
        {
            return;
        }

        refreshToken.RevokedAt = DateTime.UtcNow;
        await context.SaveChangesAsync();
    }

    private int GetRefreshTokenExpiryDays()
    {
        var rawValue = configuration["Jwt:RefreshTokenExpiryDays"]
            ?? throw new InvalidOperationException("RefreshTokenExpiryDays is not set");

        if (!int.TryParse(rawValue, out var tokenExpiryDays))
        {
            throw new InvalidOperationException("RefreshTokenExpiryDays is invalid");
        }

        if (tokenExpiryDays <= 0)
        {
            throw new InvalidOperationException("RefreshTokenExpiryDays must be greater than zero");
        }

        return tokenExpiryDays;
    }

    private static string CreateSecureToken()
    {
        return Convert.ToBase64String(RandomNumberGenerator.GetBytes(64));
    }

    private async Task RevokeTokenFamily(RefreshToken token)
    {
        // Reuse detection: if an old/invalid token in a chain is presented, revoke the whole family.
        var userTokens = await context.RefreshTokens
            .Where(t => t.UserId == token.UserId && !t.IsRevoked)
            .ToListAsync();

        if (userTokens.Count == 0)
        {
            return;
        }

        var now = DateTime.UtcNow;
        foreach (var userToken in userTokens)
        {
            userToken.RevokedAt = now;
        }

        logger.LogWarning("Detected refresh token reuse for user {UserId}; revoked {Count} active refresh tokens", token.UserId, userTokens.Count);

        await context.SaveChangesAsync();
    }
}
