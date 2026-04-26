namespace AuthService.Dtos;

public sealed record MeResponseDto(
    string Id,
    string Email,
    string DisplayName,
    string SchoolName,
    string Role);