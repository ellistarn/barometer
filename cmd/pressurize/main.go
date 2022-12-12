package main

import (
	"context"

	"github.com/samber/lo"
	"go.uber.org/zap"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/signals"
)

func main() {
	ctx := signals.NewContext()
	ctx = logging.WithLogger(ctx, lo.Must(zap.NewDevelopment()).Sugar())
	pressurizeCPU(ctx)
}

func pressurizeCPU(ctx context.Context) {
	for {
	}
}

func pressurizeMemory(ctx context.Context) {

}
