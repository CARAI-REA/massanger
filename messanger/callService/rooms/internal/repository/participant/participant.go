package participant

import (
	"sync"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/tokens"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	def "rooms/internal/repository"
)


var _ def.ParticipantRepository = (*participantRepository)(nil)

var (
	psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
)

type participantRepository struct {
	mu    sync.RWMutex
	pool  *pgxpool.Pool
	token tokens.JoinTokenService
}

func NewParticipantRepository(dbPool *pgxpool.Pool, tokenService tokens.JoinTokenService) *participantRepository {
	return &participantRepository{
		pool:  dbPool,
		token: tokenService,
	}
}
