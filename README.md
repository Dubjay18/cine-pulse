# Cine Pulse

![Cine Pulse Logo](https://via.placeholder.com/200x60?text=Cine+Pulse)

A sophisticated movie and series monitoring system that automatically scrapes content from various sources, processes it using AI models (Google Gemini and OpenAI), and sends beautifully formatted email notifications.

## Features

- **Automated Content Scraping**: Regularly scrapes websites for the latest movies and series information
- **AI-Powered Content Processing**: Uses Google Gemini or OpenAI to extract and analyze content data
- **Smart Scheduling**: Configurable scheduler with morning and evening jobs
- **Persistent SQLite Storage**: Reliable content storage with migrations and backup
- **Email Notifications**: Beautiful HTML emails with detailed content tables
- **Docker Support**: Containerized deployment with volume persistence

## Quick Start

### 1. Set up environment variables

```bash
cp .env.example .env
# Edit .env file and add your API keys
```

### 2. Build and run with Docker Compose

```bash
# Build and start the application
docker-compose up --build

# Run in background
docker-compose up -d --build
```

### 3. Development with SQLite browser (optional)

```bash
# Include SQLite web browser for database management
docker-compose --profile dev up --build
```

Access SQLite browser at: http://localhost:8081

## Manual Docker Build

```bash
# Build the image
docker build -t cine-pulse .

# Run the container
docker run -d \
  --name cine-pulse \
  -e GEMINI_API_KEY=your_key_here \
  -v cine_pulse_data:/data \
  -p 8080:8080 \
  cine-pulse
```

## Migration Management

### Local Migration Commands

```bash
# Run all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Show migration status
make migrate-status

# Show current database version
make migrate-version

# Reset database (WARNING: deletes all data)
make migrate-reset
```

### Docker Migration Commands

```bash
# Run migrations in Docker container
make docker-migrate-up

# Check migration status in container
make docker-migrate-status

# Get database version in container
make docker-migrate-version
```

### Creating New Migrations

1. Create a new migration file in `storage/migrations/`:
```bash
# Format: YYYYMMDDHHMMSS_description.sql
touch storage/migrations/20250820000003_add_new_feature.sql
```

2. Add SQL for both up and down migrations:
```sql
-- +goose Up
ALTER TABLE content ADD COLUMN new_field TEXT;
CREATE INDEX idx_content_new_field ON content(new_field);

-- +goose Down
DROP INDEX idx_content_new_field;
-- Note: SQLite doesn't support DROP COLUMN directly
```

3. Run the migration:
```bash
make migrate-up
```

### Migration Best Practices

- **Always test migrations**: Run on development data first
- **Write rollback scripts**: Include `-- +goose Down` sections
- **Use transactions**: Goose automatically wraps migrations in transactions
- **Backup before major changes**: Use `make backup` before schema changes
- **SQLite limitations**: Be aware that SQLite doesn't support all ALTER TABLE operations

## Scheduler

The application includes a scheduler that automatically runs content scraping jobs at 10:00 AM and 5:00 PM every day.

### Email Notifications

The application sends rich, beautifully formatted email notifications with details of newly scraped content after each scraping job runs.

### Email Features

- **Rich HTML Templates**: Beautifully styled content presentation
- **Content Categorization**: Separate sections for movies and series
- **Detailed Information**: Includes titles, years, categories, and extra info
- **Responsive Design**: Looks great on desktop and mobile devices
- **Plain Text Fallback**: Compatible with all email clients
- **Rating Information**: Shows ratings when available
- **Source Attribution**: Lists all sources that were scraped

### Sample Email Format

![Email Sample](https://via.placeholder.com/500x300?text=Email+Template+Sample)

### Setting up with Mailtrap

This application uses Mailtrap for reliable email delivery. To set up Mailtrap:

1. [Create a free Mailtrap account](https://mailtrap.io/register/signup)
2. Verify your sending domain in Mailtrap
3. Navigate to Sending Domains → Integration and select Transactional Stream
4. Copy your SMTP credentials

Then set the following environment variables:
```bash
EMAIL_SMTP_HOST=live.smtp.mailtrap.io  # Mailtrap SMTP server
EMAIL_SMTP_PORT=587                    # Mailtrap SMTP port
EMAIL_SENDER=your_verified@domain.com  # Your verified sender email in Mailtrap
EMAIL_PASSWORD=your_mailtrap_password  # Mailtrap password or API token
EMAIL_RECIPIENT=you@example.com        # Recipient email address
```

For testing purposes, you can use Mailtrap's Sandbox SMTP:
```bash
EMAIL_SMTP_HOST=sandbox.smtp.mailtrap.io
EMAIL_SMTP_PORT=2525
EMAIL_USERNAME=your_sandbox_username
EMAIL_PASSWORD=your_sandbox_password
```

Email notifications are sent only when new content is successfully scraped. If no content is found or if email configuration is missing, no emails will be sent.

### Scheduler Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `RUN_MODE` | Application run mode (`scheduler` or `once`) | `scheduler` |
| `RUN_AT_STARTUP` | Run scheduled jobs at application startup | `true` |
| `SOURCE_URLS` | JSON array of URLs to scrape | `["https://nkiri.com/"]` |

### Running Modes

1. **Scheduler Mode** (default):
   ```bash
   # In docker-compose.yml
   environment:
     - RUN_MODE=scheduler
   ```
   This runs the application as a daemon that executes jobs at scheduled times (10 AM and 5 PM).

2. **Single Run Mode**:
   ```bash
   # In docker-compose.yml
   environment:
     - RUN_MODE=once
   ```
   This runs the scraping job once and then exits.

### Custom Scraper Sources

You can specify custom sources to scrape by setting the `SOURCE_URLS` environment variable:

```bash
# In docker-compose.yml
environment:
  - SOURCE_URLS=["https://nkiri.com/", "https://example.com/movies"]
```

## Database Management

### View database file location
The SQLite database is stored in a Docker volume at `/data/cine_pulse.db`
```bash
# Copy database from container
docker cp cine-pulse-app:/data/cine_pulse.db ./backup.db
```

### Restore database
```bash
# Copy database to container
docker cp ./backup.db cine-pulse-app:/data/cine_pulse.db
```

### Access database directly
```bash
# Enter container
docker exec -it cine-pulse-app sh

# Or with SQLite CLI (if installed)
docker exec -it cine-pulse-app sqlite3 /data/cine_pulse.db
```

## Storage Features

- **Database migrations with Goose**: Version-controlled schema changes
- **Automatic migration on startup**: Database schema is automatically updated
- **Migration CLI tool**: Manual migration management
- **Content deduplication**: Uses INSERT OR REPLACE to avoid duplicates
- **Indexed searches**: Optimized queries with database indexes
- **Statistics**: Built-in content statistics (total, movies, series)
- **Search functionality**: Search content by title
- **Type filtering**: Filter content by type (movie/series)
- **Rating and source tracking**: Enhanced content metadata

## AI Content Processing

Cine-Pulse leverages powerful AI models to process and extract structured information from scraped web content.

### Supported AI Models

- **Google Gemini**: Primary model for content extraction and analysis
- **OpenAI**: Alternative model for content processing

### AI Processing Pipeline

1. **Content Scraping**: Web content is scraped from configured sources
2. **Preprocessing**: Raw HTML is cleaned and prepared for AI analysis
3. **AI Prompting**: Models are prompted to extract structured data
4. **JSON Extraction**: Content details are extracted from model responses
5. **Data Validation**: Extracted data is validated for consistency
6. **Manual Fallback**: If JSON parsing fails, regex patterns extract data
7. **Storage**: Processed content is stored in the database

### AI-Powered Features

- **Intelligent Title Extraction**: Identifies movie and series titles
- **Year Detection**: Extracts release years when available
- **Category Classification**: Categorizes content by genre
- **Rating Recognition**: Extracts numerical ratings when available
- **Extra Info Parsing**: Captures additional relevant details

### AI Configuration

Configure your preferred model using environment variables:
```bash
# For Google Gemini
GEMINI_API_KEY=your_gemini_api_key

# For OpenAI (optional)
OPENAI_API_KEY=your_openai_api_key
```

## Project Structure

```
cine-pulse/
├── storage/
│   ├── base.go              # Content struct definition
│   ├── sqlite.go            # SQLite storage implementation
│   ├── migrations.go        # Goose migration manager
│   └── migrations/          # Database migration files
│       ├── 20250820000001_initial_schema.sql
│       └── 20250820000002_add_rating_and_source.sql
├── cmd/
│   ├── main.go              # Application entry point
│   ├── migrate/             # Migration CLI tool
│   │   └── main.go
│   └── test_email/          # Email testing utility
│       └── main.go
├── model/                   # AI model integrations
│   ├── gemini.go            # Google Gemini implementation
│   ├── manager.go           # Model manager
│   ├── model.go             # Model interfaces
│   └── openai.go            # OpenAI implementation
├── scraper/                 # Web scraping logic
│   └── scraper.go           # Scraper implementation
├── notifier/                # Notification system
│   └── email.go             # Email notification implementation
├── scheduler/               # Job scheduler
│   ├── scheduler.go         # Cron job scheduler
│   └── content_scraper_job.go # Content scraping job implementation  
├── Dockerfile               # Docker build configuration
├── docker-compose.yml       # Docker Compose configuration
├── .env.example             # Environment variables template
└── .dockerignore            # Docker build ignore rules
```

## Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| **AI Models** | | | |
| `GEMINI_API_KEY` | Google Gemini API key | Yes | - |
| `OPENAI_API_KEY` | OpenAI API key | Optional | - |
| **Application Settings** | | | |
| `DATA_PATH` | Database storage path | No | `/data` |
| `LOG_LEVEL` | Logging level | No | `info` |
| `PORT` | Application port | No | `8080` |
| `RUN_MODE` | Application run mode (`scheduler` or `once`) | No | `scheduler` |
| `RUN_AT_STARTUP` | Run scheduled jobs at application startup | No | `true` |
| **Content Sources** | | | |
| `SOURCE_URLS` | JSON array of URLs to scrape | No | `["https://nkiri.com/"]` |
| **Email Notification** | | | |
| `EMAIL_SMTP_HOST` | SMTP server hostname | For email | - |
| `EMAIL_SMTP_PORT` | SMTP server port | No | `587` |
| `EMAIL_SENDER` | Sender email address | For email | - |
| `EMAIL_PASSWORD` | Password or API token | For email | - |
| `EMAIL_RECIPIENT` | Recipient email address | For email | - |
| `EMAIL_USERNAME` | SMTP username (for testing) | Optional | - |

## Docker Commands Reference

```bash
# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Remove volumes (WARNING: deletes database)
docker-compose down -v

# Rebuild without cache
docker-compose build --no-cache

# View database volume
docker volume inspect cine-pulse_cine_pulse_data
```

## Health Check

The container includes a health check that verifies the database file exists:

```bash
# Check container health
docker ps
# Look for "healthy" status
```

## Troubleshooting

### Database permissions
If you encounter permission issues:
```bash
# Check container logs
docker-compose logs cine-pulse

# Ensure volume permissions
docker exec -it cine-pulse-app ls -la /data
```

### Build issues
If build fails with SQLite errors:
```bash
# Clean build
docker-compose down
docker system prune -f
docker-compose build --no-cache
```

### CGO requirements
This project requires CGO for SQLite driver. The Dockerfile handles this automatically with:
- gcc and musl-dev for building
- CGO_ENABLED=1 during build

## Email Troubleshooting

### Mailtrap Email Sending Issues

#### Authentication Failed
If you see `Failed to send email: failed to send email: 535 5.7.8 Authentication failed`:
1. Check that you're using `api` as the username when using an API token
2. Verify that your API token is correct and has the "Send Emails" permission
3. Check that your token is correctly formatted in the .env file without any spaces

#### Domain Credibility Check
If you see `550 5.7.1 Security check pending. Mailtrap is checking your domain credibility`:
1. Use a sender email from a domain you've verified in Mailtrap
2. Alternatively, use `no-reply@smtp.mailtrap.io` as your sender
3. For testing, switch to Mailtrap's sandbox environment instead

#### Testing Email Configuration
Use the test email utility to verify your configuration:

```bash
# Run the email test utility
go run cmd/test_email/main.go
```

This will attempt to send a test email using your current configuration and provide detailed logs.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
