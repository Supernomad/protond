// Copyright (c) 2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/protond/blob/master/LICENSE

package main

import (
	"os"

	"github.com/Supernomad/protond/common"
	"github.com/Supernomad/protond/filter"
	"github.com/Supernomad/protond/input"
	"github.com/Supernomad/protond/output"
	"github.com/Supernomad/protond/worker"
)

func handleError(log *common.Logger, err error) {
	if err != nil {
		log.Error.Println(err.Error())
		os.Exit(1)
	}
}

func main() {
	log := common.NewLogger(common.InfoLogger)

	cfg, err := common.NewConfig(log)
	handleError(cfg.Log, err)

	workers := make([]*worker.Worker, cfg.NumWorkers)

	filters := make([]filter.Filter, 0)
	for i := 0; i < len(cfg.Filters); i++ {
		temp, err := filter.New(cfg.Filters[i].Type, cfg.Filters[i], cfg)
		handleError(cfg.Log, err)

		filters = append(filters, temp)
	}

	if len(filters) == 0 {
		noop, _ := filter.New(filter.NoopFilter, nil, cfg)
		filters = append(filters, noop)
	}

	inputs := make([]input.Input, 0)
	for i := 0; i < len(cfg.Inputs); i++ {
		temp, err := input.New(cfg.Inputs[i].Type, cfg.Inputs[i], cfg)
		handleError(cfg.Log, err)

		err = temp.Open()
		handleError(cfg.Log, err)

		inputs = append(inputs, temp)
	}

	if len(inputs) == 0 {
		stdin, _ := input.New(input.StdinInput, nil, cfg)
		inputs = append(inputs, stdin)
	}

	stdout, _ := output.New(output.StdoutOutput, cfg)

	for i := 0; i < cfg.NumWorkers; i++ {
		workers[i] = worker.New(cfg, inputs, filters, []output.Output{stdout})
		workers[i].Start()
	}

	signaler := common.NewSignaler(log, cfg, nil, map[string]string{})

	log.Info.Println("[MAIN]", "protond start up complete.")

	err = signaler.Wait(true)
	handleError(cfg.Log, err)

	for i := 0; i < cfg.NumWorkers; i++ {
		workers[i].Stop()
	}
}
