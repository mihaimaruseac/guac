FROM arigaio/atlas:latest-alpine@sha256:2f2b2744bd5e3796c9cbd034785eef00d289b6fa795e02d2a55835562efec09a

WORKDIR /app

# Copy the migration files
COPY migrations ./migrations

# Copy the atlas migration script and make it executable
COPY atlas.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]
