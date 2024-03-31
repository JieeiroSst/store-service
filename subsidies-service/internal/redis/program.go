package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/JIeerioSst/subsidies-service/dto"
	"github.com/JIeerioSst/subsidies-service/utils"
	"github.com/go-redis/redis/v8"
	"google.golang.org/appengine/log"
)

type cacheProgram struct {
	redis *redis.Client
}

const (
	programHashKey       = "program-hash-key-program-id-%v"
	programHashPagingKey = "program-hash-key-program-id-"
)

type CacheProgram interface {
	SetProgramCache(ctx context.Context, programs []dto.Program) error
	UpdateDiscountsProgramByIdCache(ctx context.Context, programID int, discounts []dto.Discount) error
	GetProgramCache(ctx context.Context, ids []int) ([]dto.Program, map[int]bool, error)
	GetProgramsWithPaging(ctx context.Context, page int, limit int) (*dto.ProgramPage, error)
}

func NewCacheProgram(redis *redis.Client) CacheProgram {
	return &cacheProgram{redis: redis}
}

func (c *cacheProgram) SetProgramCache(ctx context.Context, programs []dto.Program) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if duration.Seconds() >= 1 {
			log.Infof(ctx, "Slow request SetProgramCache in seconds = %v", duration.Seconds())
		}
	}()

	pipe := c.redis.Pipeline()
	for _, program := range programs {
		pipe.HMSet(ctx, fmt.Sprintf(programHashKey, program.ID), utils.ArgsParameter{}.AddFlat(program).ParseJson()...)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Errorf(ctx, "SetProgramCache in err = %v", err)
		return err
	}
	return nil
}

func (c *cacheProgram) UpdateDiscountsProgramByIdCache(ctx context.Context, programID int, discounts []dto.Discount) error {
	discountsJson, _ := json.Marshal(&discounts)
	err := c.redis.HMSet(ctx, fmt.Sprintf(programHashKey, programID), map[string]interface{}{"discounts": discountsJson}).Err()
	if err != nil {
		log.Errorf(ctx, "SetProgramCache in err = %v", err)
		return err
	}
	return nil
}

func (c *cacheProgram) GetProgramCache(ctx context.Context, ids []int) (result []dto.Program, idsExist map[int]bool, err error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if duration.Seconds() >= 1 {
			log.Infof(ctx, "Slow request GetProgramCache in seconds = %v", duration.Seconds())
		}
	}()
	var listCmd []*redis.StringStringMapCmd
	pipe := c.redis.Pipeline()
	for _, id := range ids {
		listCmd = append(listCmd, pipe.HGetAll(ctx, fmt.Sprintf(programHashKey, id)))
	}
	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		log.Errorf(ctx, "GetProgramCache in err = %v", err)
		return nil, nil, err
	}
	idsExist = map[int]bool{}
	for _, cmd := range listCmd {
		var program dto.Program
		res, err := cmd.Result()
		if err != nil && err != redis.Nil {
			continue
		}
		if len(res["discounts"]) > 0 {
			json.Unmarshal([]byte(res["discounts"]), &program.Discounts)
		}
		err = cmd.Scan(&program)
		if err != nil {
			continue
		}

		if program.ID > 0 {
			idsExist[program.ID] = true
			result = append(result, program)
		}
	}
	return result, idsExist, nil
}

// limit := 10
// page := 0 first page
func (c *cacheProgram) GetProgramsWithPaging(ctx context.Context, page int, limit int) (*dto.ProgramPage, error) {
	keys, nextPage, err := c.redis.HScan(ctx, programHashPagingKey, uint64(page), "", int64(limit)).Result()
	if err != nil {
		return nil, err
	}

	programMap := make(map[string]map[string]string)
	for _, key := range keys {
		programData, err := c.redis.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		programMap[key] = programData
	}
	programs := make([]dto.Program, 0, len(programMap))
	for _, data := range programMap {
		var program dto.Program
		err := json.Unmarshal([]byte(data["program"]), &program)
		if err != nil {
			return nil, err
		}
		programs = append(programs, program)
	}
	hasNext := nextPage != 0

	return &dto.ProgramPage{Programs: programs, HasNext: hasNext}, nil
}
