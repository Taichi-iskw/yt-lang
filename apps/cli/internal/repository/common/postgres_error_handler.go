package repository

import (
	"errors"
	"strings"

	apperrors "github.com/Taichi-iskw/yt-lang/internal/errors"
	"github.com/jackc/pgx/v5/pgconn"
)

// handlePostgreSQLError converts PostgreSQL-specific errors to appropriate AppError codes
func handlePostgreSQLError(err error, operation string) *apperrors.AppError {
	if err == nil {
		return nil
	}

	// Check if it's a PostgreSQL error
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		// Not a PostgreSQL error, return generic internal error
		return apperrors.Wrap(err, apperrors.CodeInternal, operation)
	}

	// Map PostgreSQL error codes to AppError codes
	switch pgErr.Code {
	case "23505": // UNIQUE_VIOLATION
		return handleUniqueViolation(pgErr, operation)

	case "23503": // FOREIGN_KEY_VIOLATION
		return handleForeignKeyViolation(pgErr, operation)

	case "23502": // NOT_NULL_VIOLATION
		return apperrors.Wrap(err, apperrors.CodeInvalidArg, "required field is missing")

	case "23514": // CHECK_VIOLATION
		return apperrors.Wrap(err, apperrors.CodeInvalidArg, "data violates check constraint")

	case "42P01": // UNDEFINED_TABLE
		return apperrors.Wrap(err, apperrors.CodeInternal, "database schema error: table not found")

	case "42703": // UNDEFINED_COLUMN
		return apperrors.Wrap(err, apperrors.CodeInternal, "database schema error: column not found")

	case "08000", "08003", "08006": // CONNECTION_EXCEPTION variants
		return apperrors.Wrap(err, apperrors.CodeInternal, "database connection error")

	case "53300": // TOO_MANY_CONNECTIONS
		return apperrors.Wrap(err, apperrors.CodeInternal, "database connection limit reached")

	default:
		// Unknown PostgreSQL error, return with error code for debugging
		message := "database error (PostgreSQL code: " + pgErr.Code + ")"
		return apperrors.Wrap(err, apperrors.CodeInternal, message)
	}
}

// handleUniqueViolation provides specific error messages for different unique constraints
func handleUniqueViolation(pgErr *pgconn.PgError, operation string) *apperrors.AppError {
	constraintName := pgErr.ConstraintName

	// Provide user-friendly messages based on constraint
	switch {
	case strings.Contains(constraintName, "pkey"):
		// Primary key violation
		if strings.Contains(constraintName, "channels") {
			return apperrors.Wrap(pgErr, apperrors.CodeConflict, "channel with this ID already exists")
		} else if strings.Contains(constraintName, "videos") {
			return apperrors.Wrap(pgErr, apperrors.CodeConflict, "video with this ID already exists")
		}
		return apperrors.Wrap(pgErr, apperrors.CodeConflict, "resource with this ID already exists")

	case strings.Contains(constraintName, "url"):
		// URL unique constraint
		if strings.Contains(constraintName, "channels") {
			return apperrors.Wrap(pgErr, apperrors.CodeConflict, "channel with this URL already exists")
		} else if strings.Contains(constraintName, "videos") {
			return apperrors.Wrap(pgErr, apperrors.CodeConflict, "video with this URL already exists")
		}
		return apperrors.Wrap(pgErr, apperrors.CodeConflict, "resource with this URL already exists")

	default:
		// Generic unique violation
		return apperrors.Wrap(pgErr, apperrors.CodeConflict, "resource already exists")
	}
}

// handleForeignKeyViolation provides specific error messages for foreign key constraints
func handleForeignKeyViolation(pgErr *pgconn.PgError, operation string) *apperrors.AppError {
	constraintName := pgErr.ConstraintName

	// Provide user-friendly messages based on foreign key constraint
	switch {
	case strings.Contains(constraintName, "channel_id"):
		return apperrors.Wrap(pgErr, apperrors.CodeDependency, "referenced channel does not exist")

	case strings.Contains(constraintName, "video_id"):
		return apperrors.Wrap(pgErr, apperrors.CodeDependency, "referenced video does not exist")

	default:
		// Generic foreign key violation
		return apperrors.Wrap(pgErr, apperrors.CodeDependency, "referenced resource does not exist")
	}
}
