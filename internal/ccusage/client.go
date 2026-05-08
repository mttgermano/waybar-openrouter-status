package ccusage

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
 	"time"
)

type TokenCounts struct {
	InputTokens              int `json:"inputTokens"`
	OutputTokens             int `json:"outputTokens"`
	CacheCreationInputTokens int `json:"cacheCreationInputTokens"`
	CacheReadInputTokens     int `json:"cacheReadInputTokens"`
}


type Block struct {
	ID            string      `json:"id"`
 	EndTime       string      `json:"endTime"`
	IsActive      bool        `json:"isActive"`
	Entries       int         `json:"entries"`
	TokenCounts   TokenCounts `json:"tokenCounts"`
	TotalTokens   int         `json:"totalTokens"`
	CostUSD       float64     `json:"costUSD"`
	Models        []string    `json:"models"`
}

type BlocksResponse struct {
	Blocks []Block `json:"blocks"`
}

type BlocksData struct {
	Entries                  	int
	TotalTokens              	int
	BlockTotalTokens			int
	InputTokens              	int
	OutputTokens             	int
	CostUSD                  	float64
	//CostPerHour              	float64
	LastModel                	string
 	EndTime       				string
}

type DailyResponse struct {
	Daily []Daily `json:"daily"`
}

type Daily struct {
	TotalTokens 			int `json:"totalTokens"`
	InputTokens 			int `json:"inputTokens"`
	OutputTokens 			int `json:"outputTokens"`
	CacheCreationTokens 	int `json:"cacheCreationTokens"`
	CacheReadTokens 		int `json:"cacheReadTokens"`
	Models        			[]string `json:"modelsUsed"`
}

type DailyData struct {
	TotalTokens            	int
	InputTokens				int
	OutputTokens        	int
	LastModel				string
}

type Data struct {
	BlockInputTokens        int
	BlockOutputTokens       int
	BlockTotalTokens        int
	DailyTotalTokens        int
	DailyOutputTokens       int
	DailyInputTokens        int
	Entries					int
	LastModel               string
}


func GetData(ctx context.Context) (*Data, error) {
	block, err := getBlocks(ctx)
	if err != nil {
		return nil, err
	}
	daily, err := getDaily(ctx)
	if err != nil {
		return nil, err
	}

	lastModel := block.LastModel
	if lastModel == "" {
		lastModel = daily.LastModel
	}

	return &Data{
		BlockTotalTokens:   block.TotalTokens,
		BlockInputTokens:  	block.InputTokens,
		BlockOutputTokens: 	block.OutputTokens,
		LastModel:         	lastModel,
		DailyInputTokens:  	daily.InputTokens,
		DailyOutputTokens: 	daily.OutputTokens,
		DailyTotalTokens:  	daily.TotalTokens,
		Entries:          	block.Entries,
	}, nil
}

func getDaily(ctx context.Context) (*DailyData,error) {
	cmd := exec.CommandContext(ctx, "npx", "ccusage@latest", "daily", "--active", "--json", "--offline")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("execute ccusage (npx ccusage@latest daily --active --json --offline): %w", err)
	}

	var response DailyResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("parse ccusage json output: %w", err)
	}

	if len(response.Daily) == 0 {
		return nil, fmt.Errorf("no active usage daily found in ccusage response")
	}

	daily := response.Daily[0]

	lastModel := ""
	if len(daily.Models) > 0 {
		lastModel = daily.Models[len(daily.Models)-1]
	}

	return &DailyData{
		TotalTokens: 	daily.TotalTokens,
		InputTokens:  	daily.InputTokens + daily.CacheCreationTokens + daily.CacheReadTokens,
		OutputTokens:   daily.OutputTokens,
		LastModel:		lastModel,
	}, nil
}

func countEntries(blocks BlocksResponse) (int) {
	now := time.Now().UTC()
	var sum int

	for _, b := range blocks.Blocks {
		t, err := time.Parse(time.RFC3339, b.EndTime)
		if err != nil {
			continue
		}

		if now.Sub(t) <= 24*time.Hour && now.After(t) {
			sum += b.Entries
		}
	}
	return sum
}


func getBlocks(ctx context.Context) (*BlocksData, error) {
	cmd := exec.CommandContext(ctx, "npx", "ccusage@latest", "blocks", "--json", "--offline")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("execute ccusage (npx ccusage@latest blocks --json --offline): %w", err)
	}

	var response BlocksResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("parse ccusage json output: %w", err)
	}

	block := Block{}
	if len(response.Blocks) == 0 {
		block = Block{
			ID:          "0",
			Entries:     0,
			TokenCounts: TokenCounts{},
			TotalTokens: 0,
			CostUSD:     0,
			Models:      []string{},
		}
	} else {
		block = response.Blocks[0]
	}

	lastModel := ""
	if len(block.Models) > 0 {
		lastModel = block.Models[len(block.Models)-1]
	}

	return &BlocksData{
		Entries:          	countEntries(response),
		TotalTokens:  		block.TotalTokens,
		InputTokens:        block.TokenCounts.InputTokens + block.TokenCounts.CacheCreationInputTokens + block.TokenCounts.CacheReadInputTokens,
		OutputTokens:       block.TokenCounts.OutputTokens,
		LastModel:          lastModel,
		EndTime: 			block.EndTime,
	}, nil
}
