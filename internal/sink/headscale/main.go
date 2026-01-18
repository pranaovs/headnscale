package headscale

import (
	"context"
	"log"

	"github.com/pranaovs/headnscale/internal/config"
	"github.com/pranaovs/headnscale/internal/types"
)

func New(config config.Config) *Sink {
	return &Sink{
		Common:    config.Common,
		Headscale: config.Sink.Headscale,
	}
}

func (s *Sink) Initialize(ctx context.Context) error {
	log.Printf("Headscale sink initialized")
	return nil
}

func (s *Sink) Process(ctx context.Context, nodes []types.Node) error {
	records := create(nodes, s.BaseDomain)
	if s.NoBaseDomain {
		records = append(records, create(nodes, "")...)
	}
	sorted := sort(records)

	if err := write(s.ExtraRecordsFile, sorted); err != nil {
		log.Printf("error writing JSON: %v", err)
		return err
	}

	log.Printf("Wrote %d DNS records to %s", len(sorted), s.ExtraRecordsFile)
	return nil
}

func (s *Sink) Close() error {
	return nil
}
