package job

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/C0deNe0/go-boiler/internal/config"
	"github.com/C0deNe0/go-boiler/internal/lib/email"
	"github.com/hibiken/asynq"

	// "github.com/hibiken/asynq"
	"github.com/rs/zerolog"
)

var emailClient *email.Client

func (j *JobService) InitHandlers(config *config.Config, logger *zerolog.Logger) {
	emailClient = email.NewClient(logger, config)
}

func (j *JobService) handleWelcomeEmailTask(ctx context.Context, t *asynq.Task) error {
	var p WelcomeEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshall incoming email payload: %w", err)
	}

	j.logger.Info().Str("type", "welcome").Str("to", p.To).Msg("Processing welcome email task")

	err := emailClient.SendWelcomeEmail(
		p.To,
		p.FirstName,
	)

	if err != nil {
		j.logger.Info().Str("type", "welcome").Err(err).Msg("Failed to send welcome email")
		return err
	}

	j.logger.Info().Str("type", "welcome").Str("to", p.To).Msg("successfully send welcome email")
	return nil
}
