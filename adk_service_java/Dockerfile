# Multi-stage Docker build for ADK Widget Builder Service

# Build stage
FROM maven:3.9-eclipse-temurin-17 AS builder

WORKDIR /app

# Copy POM file first for better caching
COPY pom.xml .

# Download dependencies
RUN mvn dependency:go-offline -B

# Copy source code
COPY src ./src

# Build the application
RUN mvn clean package -DskipTests

# Runtime stage
FROM eclipse-temurin:17-jre

# Install curl for health checks
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the built JAR from builder stage
COPY --from=builder /app/target/widget-builder-adk-*.jar app.jar

# Create non-root user
RUN groupadd -r adkuser && useradd -r -g adkuser adkuser
RUN chown -R adkuser:adkuser /app
USER adkuser

# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD curl -f http://localhost:8081/api/adk/health || exit 1

# Run the application
ENTRYPOINT ["java", "-jar", "/app/app.jar"]

# JVM tuning for container environment
ENV JAVA_OPTS="-XX:+UseContainerSupport -XX:MaxRAMPercentage=75.0 -XX:+UseG1GC"
CMD ["sh", "-c", "java $JAVA_OPTS -jar /app/app.jar"]