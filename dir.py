from pathlib import Path

ROOT = Path(".")

directories = [
    # =========================
    # Commands
    # =========================
    "cmd/api",
    "cmd/seed",

    # =========================
    # Internal
    # =========================
    "internal",

    # Application bootstrap
    "internal/app",

    # =========================
    # HTTP Layer
    # =========================
    "internal/api",
    "internal/api/handlers",
    "internal/api/middleware",
    "internal/api/response",

    # =========================
    # Platform (Infrastructure)
    # =========================
    "internal/platform",

    "internal/platform/config",
    "internal/platform/database",
    "internal/platform/logger",
    "internal/platform/telemetry",
    "internal/platform/migrate",
    "internal/platform/auth",
    "internal/platform/storage",
    "internal/platform/queue",
    "internal/platform/nim",
    "internal/platform/embedding",

    # =========================
    # Domain Packages
    # =========================
    "internal/catalog",
    "internal/search",
    "internal/features",
    "internal/ai",
    "internal/users",
    "internal/cart",
    "internal/orders",
    "internal/collections",
    "internal/chat",
    "internal/notifications",

    # =========================
    # Background Jobs
    # =========================
    "internal/jobs",
    "internal/jobs/catalog",
    "internal/jobs/reviews",
    "internal/jobs/features",
    "internal/jobs/embeddings",
    "internal/jobs/summaries",

    # =========================
    # Database
    # =========================
    "db",
    "db/migrations",
    "db/queries",

    # =========================
    # Assets
    # =========================
    "assets",
    "assets/prompts",
    "assets/schemas",

    # =========================
    # Configuration
    # =========================
    "configs",

    # =========================
    # Deployment
    # =========================
    "deploy",
    "deploy/docker",

    # =========================
    # Scripts
    # =========================
    "scripts",

    # =========================
    # Documentation
    # =========================
    "docs",
    "docs/api",
    "docs/architecture",

    # =========================
    # Tests
    # =========================
    "tests",
    "tests/integration",
    "tests/testdata",
]

for directory in directories:
    (ROOT / directory).mkdir(parents=True, exist_ok=True)

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

print(f"✅ Created {len(directories)} directories")
print("✅ Added .gitkeep files")