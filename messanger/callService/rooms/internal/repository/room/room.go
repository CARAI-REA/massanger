package room

import (
	"sync"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/tokens"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	def "rooms/internal/repository"
)


var _ def.RoomRepository = (*roomRepository)(nil)

var (
	psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
)

type roomRepository struct {
	mu    sync.RWMutex
	pool  *pgxpool.Pool
	token tokens.JoinTokenService
}

func NewRoomRepository(dbPool *pgxpool.Pool, tokenService tokens.JoinTokenService) *roomRepository {
	return &roomRepository{
		pool:  dbPool,
		token: tokenService,
	}
}
