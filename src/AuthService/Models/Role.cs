namespace AuthService.Models;

public class Role
{
    public int Id { get; set; }
    public required string Name { get; set; }
    public required string Code { get; set; }
    public ICollection<User> Users { get; set; } = new List<User>();

}