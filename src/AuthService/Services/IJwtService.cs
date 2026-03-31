using AuthService.Models;

namespace AuthService.Services;

public interface IJwtService
{
    Task<string> GenerateJwtToken(User user);
}
