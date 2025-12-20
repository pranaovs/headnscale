package headscale

import (
	"context"
	"log"

	"github.com/pranaovs/headnscale/internal/types"
)

func New(config types.Config) *Sink {
	return &Sink{
		filePath:     config.ExtraRecordsFile,
		noBaseDomain: config.NoBaseDomain,
		baseDomain:   config.BaseDomain,
	}
}

func (s *Sink) Initialize(ctx context.Context) error {
	log.Printf("Headscale sink initialized")
	return nil
}

func (s *Sink) Process(ctx context.Context, nodes []types.Node) error {
	records := create(nodes, s.baseDomain)
	if s.noBaseDomain {
		records = append(records, create(nodes, "")...)
	}
	sorted := sort(records)

	if err := write(s.filePath, sorted); err != nil {
		log.Printf("error writing JSON: %v", err)
		return err
	}

	log.Printf("Wrote %d DNS records to %s", len(sorted), s.filePath)
	return nil
}

func (s *Sink) Close() error {
	return nil
}
