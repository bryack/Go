# Requirements Document

## Introduction

This document defines the requirements for implementing a comprehensive configuration management system for the Task Manager Server using command-line flags and Viper. The system will support flexible configuration sources including flags, environment variables, and configuration files.

## Glossary

- **Server**: The HTTP API server that handles task management operations
- **Viper**: A Go configuration library that supports multiple configuration sources
- **Flag**: Command-line argument passed when starting the server application
- **Config File**: A YAML or JSON file containing configuration settings
- **Environment Variable**: System-level variable that can override configuration values
- **Configuration Precedence**: The order in which configuration sources are evaluated (flags > env vars > config file > defaults)

## Requirements

### Requirement 1: Server Configuration Management

**User Story:** As a server administrator, I want to configure the server port, database path, JWT settings, and other parameters through multiple methods, so that I can deploy the server in different environments

#### Acceptance Criteria

1. WHEN the Server starts, THE Server SHALL load configuration from flags, environment variables, config file, and defaults in that precedence order
2. WHERE a `--port` flag is provided, THE Server SHALL listen on the specified port
3. WHERE a `--db-path` flag is provided, THE Server SHALL use the specified database file path
4. WHERE a `--jwt-secret` flag is provided, THE Server SHALL use the specified JWT secret key
5. WHERE a `--jwt-expiration` flag is provided, THE Server SHALL use the specified token expiration duration
6. THE Server SHALL validate all configuration values before starting

### Requirement 2: Configuration File Support

**User Story:** As a server administrator, I want to store my configuration in a file, so that I don't have to specify settings every time I run the server

#### Acceptance Criteria

1. THE System SHALL support YAML format for configuration files
2. THE System SHALL support JSON format for configuration files
3. WHEN a configuration file is found, THE System SHALL parse and validate all settings
4. IF a configuration file contains invalid syntax, THEN THE System SHALL return a descriptive error message
5. THE System SHALL support nested configuration structures for organized settings

### Requirement 3: Environment Variable Support

**User Story:** As a DevOps engineer, I want to configure the application using environment variables, so that I can easily deploy in containerized environments

#### Acceptance Criteria

1. THE System SHALL support environment variables with a consistent prefix (`TASKMANAGER_`)
2. THE System SHALL automatically map environment variables to configuration keys (e.g., `TASKMANAGER_SERVER_URL` maps to `server.url`)
3. WHEN an environment variable is set, THE System SHALL override config file values but be overridden by flags
4. THE System SHALL support both underscore and dot notation for nested configuration keys

### Requirement 4: Configuration Validation

**User Story:** As a server administrator, I want clear error messages when my configuration is invalid, so that I can quickly fix configuration issues

#### Acceptance Criteria

1. WHEN the Server URL is configured, THE System SHALL validate it is a valid HTTP or HTTPS URL
2. WHEN the Server port is configured, THE System SHALL validate it is between 1 and 65535
3. WHEN the database path is configured, THE System SHALL validate the directory exists and is writable
4. WHEN the JWT expiration is configured, THE System SHALL validate it is a positive duration
5. IF any validation fails, THEN THE System SHALL return a descriptive error message indicating which setting is invalid

### Requirement 5: Default Configuration Values

**User Story:** As a server administrator, I want sensible default values for all settings, so that I can run the server without extensive configuration

#### Acceptance Criteria

1. THE Server SHALL default to port `8080` for HTTP listening
2. THE Server SHALL default to `./data/tasks.db` for the database path
3. THE Server SHALL default to `24h` for JWT token expiration
4. THE Server SHALL document all default values in help text and documentation

### Requirement 6: Configuration Display and Help

**User Story:** As a server administrator, I want to see available configuration options and their current values, so that I can understand how to configure the server

#### Acceptance Criteria

1. WHEN the user runs the server with `--help` flag, THE Server SHALL display all available configuration flags with descriptions
2. WHERE a `--show-config` flag is provided, THE Server SHALL display the current effective configuration with sources
3. THE Server SHALL indicate which configuration source provided each value (flag, env, file, default)
4. THE Server SHALL display configuration in a human-readable format

### Requirement 7: Configuration Precedence

**User Story:** As a server administrator, I want a clear and predictable order for configuration sources, so that I can override settings as needed

#### Acceptance Criteria

1. THE Server SHALL apply configuration in the following precedence order: command-line flags (highest), environment variables, configuration file, defaults (lowest)
2. WHEN multiple configuration sources provide the same setting, THE Server SHALL use the value from the highest precedence source
3. THE Server SHALL document the precedence order in help text and documentation

### Requirement 8: Backward Compatibility

**User Story:** As an existing server administrator, I want the new configuration system to work with my current setup, so that I don't have to change my deployment

#### Acceptance Criteria

1. THE Server SHALL continue to support the existing `JWT_SECRET_KEY` environment variable
2. WHEN the legacy `JWT_SECRET_KEY` environment variable is used, THE Server SHALL map it to the new configuration structure
3. THE Server SHALL log a deprecation warning when legacy environment variables are detected

### Requirement 9: Configuration Security

**User Story:** As a security-conscious administrator, I want sensitive configuration values to be handled securely, so that credentials are not exposed

#### Acceptance Criteria

1. WHEN displaying configuration with `--show-config`, THE Server SHALL mask sensitive values (JWT secret, passwords)
2. THE Server SHALL not log sensitive configuration values in plain text
3. WHEN a configuration file contains sensitive values, THE Server SHALL warn if file permissions are too permissive (world-readable)
4. THE Server SHALL support reading sensitive values from environment variables instead of files
