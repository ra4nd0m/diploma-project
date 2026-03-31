using AuthService.Models;

namespace AuthService.Services;

public interface IRefreshTokenService
{
    Task<RefreshToken> GenerateRefreshToken(User user);
    Task<RefreshToken?> ValidateRefreshToken(string token);
    Task<RefreshToken?> RotateRefreshToken(string token);
    Task RecallRefreshToken(string token);
}
