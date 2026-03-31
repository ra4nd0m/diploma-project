using Microsoft.AspNetCore.Identity;

namespace AuthService.Models;

public class User : IdentityUser
{
    public required string DisplayName { get; set; }
    public required string SchoolName { get; set; }
    public int RoleId { get; set; }
    public Role Role { get; set; } = null!;
    public ICollection<RefreshToken> RefreshTokens { get; set; } = new List<RefreshToken>();
}