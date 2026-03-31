namespace AuthService.Dtos;

public sealed record RegisterDto(
    string Email,
    string Password,
    string DisplayName,
    string SchoolName,
    string Role);