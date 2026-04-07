using AuthService.Data;
using Microsoft.EntityFrameworkCore;

namespace AuthService.Utils;

public static class DatabaseInitializer
{
    public static async Task InitializeAsync(WebApplication app)
    {
        using var scope = app.Services.CreateScope();
        var services = scope.ServiceProvider;
        var logger = services.GetRequiredService<ILoggerFactory>().CreateLogger("DatabaseInitializer");
        var context = services.GetRequiredService<AuthDbContext>();

        logger.LogInformation("Applying database migrations");
        await context.Database.MigrateAsync();
        logger.LogInformation("Database migrations applied");
    }
}
