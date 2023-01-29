package modules

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p/core/host"

	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/node/modules/helpers"
	"github.com/filecoin-project/lotus/node/modules/lp2p"
	"github.com/libp2p/go-libp2p"
	"go.uber.org/fx"

	"github.com/filecoin-project/lily/chain/indexer/distributed"

	"github.com/filecoin-project/lily/config"
	"github.com/filecoin-project/lily/storage"
)

func NewStorageCatalog(mctx helpers.MetricsCtx, lc fx.Lifecycle, cfg *config.Conf) (*storage.Catalog, error) {
	return storage.NewCatalog(cfg.Storage)
}

func LoadConf(path string) func(mctx helpers.MetricsCtx, lc fx.Lifecycle) (*config.Conf, error) {
	return func(mctx helpers.MetricsCtx, lc fx.Lifecycle) (*config.Conf, error) {
		return config.FromFile(path)
	}
}

func NewQueueCatalog(mctx helpers.MetricsCtx, lc fx.Lifecycle, cfg *config.Conf) (*distributed.Catalog, error) {
	return distributed.NewCatalog(cfg.Queue)
}

func NewHost(mctx helpers.MetricsCtx, lc fx.Lifecycle, params lp2p.P2PHostIn) (*host.Host, error) {
	pkey := params.Peerstore.PrivKey(params.ID)
	if pkey == nil {
		return nil, fmt.Errorf("missing private key for node ID: %s", params.ID.Pretty())
	}

	opts := []libp2p.Option{
		libp2p.Identity(pkey),
		libp2p.Peerstore(params.Peerstore),
		libp2p.NoListenAddrs,
		libp2p.Ping(true),
		libp2p.UserAgent("lily-" + build.UserVersion()),
	}
	for _, o := range params.Opts {
		opts = append(opts, o...)
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return h.Close()
		},
	})

	return &h, nil
}
