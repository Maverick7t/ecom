from pathlib import Path

# Root directory (current directory)
ROOT = Path(".")

directories = [
    # Commands
    "cmd/api",
    "cmd/seed",

    # Internal
    "internal",

    # Application bootstrap
    "internal/app",

    # API
    "internal/api",
    "internal/api/handlers",
    "internal/api/middleware",
    "internal/api/response",

    # Platform
    "internal/platform",
    "internal/platform/config",
    "internal/platform/database",
    "internal/platform/logger",
    "internal/platform/telemetry",
    "internal/platform/migrate",
    "internal/platform/auth",

    # Domains
    "internal/catalog",
    "internal/catalog/repository",
    "internal/catalog/service",

    "internal/search",
    "internal/search/repository",
    "internal/search/service",

    "internal/features",
    "internal/features/repository",
    "internal/features/service",

    "internal/embeddings",
    "internal/embeddings/service",

    "internal/ai",
    "internal/ai/client",
    "internal/ai/service",

    "internal/users",
    "internal/users/repository",
    "internal/users/service",

    "internal/cart",
    "internal/cart/repository",
    "internal/cart/service",

    "internal/orders",
    "internal/orders/repository",
    "internal/orders/service",

    "internal/collections",
    "internal/collections/repository",
    "internal/collections/service",

    "internal/chat",
    "internal/chat/repository",
    "internal/chat/service",

    "internal/notifications",
    "internal/notifications/repository",
    "internal/notifications/service",
    "internal/notifications/sse",

    # Jobs
    "internal/jobs",
    "internal/jobs/catalog",
    "internal/jobs/reviews",
    "internal/jobs/features",
    "internal/jobs/embeddings",
    "internal/jobs/summaries",

    # Database
    "db",
    "db/migrations",
    "db/queries",

    # Assets / seed data
    "assets",
    "assets/prompts",
    "assets/schemas",

    # Config
    "configs",

    # Scripts
    "scripts",

    # Docker
    "deploy",
    "deploy/docker",

    # Tests
    "tests",
    "tests/integration",
    "tests/testdata",

    # Documentation
    "docs",
    "docs/architecture",
    "docs/api",
]

for directory in directories:
    path = ROOT / directory
    path.mkdir(parents=True, exist_ok=True)

print(f"✅ Created {len(directories)} directories.")

# Optional placeholder files so Git tracks empty directories
gitkeep_dirs = [
    "db/migrations",
    "db/queries",
    "assets/prompts",
    "assets/schemas",
    "configs",
    "docs",
    "tests/testdata",
]

for directory in gitkeep_dirs:
    (ROOT / directory / ".gitkeep").touch(exist_ok=True)

print("✅ Added .gitkeep files.")