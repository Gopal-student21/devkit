package contract

import (
	"fmt"
	"os"

	"github.com/devkit/devkit/pkg/logger"
	"github.com/spf13/cobra"
)

// newSeedCommand wires `dev contract seed` — FK-aware mock data generation.
func newSeedCommand() *cobra.Command {
	var count int
	var seed int64
	var format string
	var out string

	cmd := &cobra.Command{
		Use:   "seed",
		Short: "Generate FK-aware mock seed data from contract",
		Long: `Generate referentially-consistent mock data from api.yaml.

Foreign keys (via $ref or naming convention like userId/order_id) point at
records that actually exist — no orphaned references (unlike naive fakers).

  dev contract seed                   — JSON, 5 records/schema -> seed.json
  dev contract seed --count 20        — 20 records per schema
  dev contract seed --format sql      — INSERT statements
  dev contract seed --seed 7 --out f.json  — deterministic output to file`,
		Run: func(cmd *cobra.Command, args []string) {
			if _, err := os.Stat("api.yaml"); err != nil {
				logger.Error("No api.yaml found. Run: dev contract init")
				os.Exit(1)
			}
			spec, err := LoadSpec("api.yaml")
			if err != nil {
				logger.Error(err.Error())
				os.Exit(1)
			}

			seeder := NewSeeder(spec, SeedOptions{Count: count, Seed: seed})
			seeder.Generate()

			var result string
			switch format {
			case "sql":
				result, err = seeder.ToSQL()
			default:
				format = "json"
				result, err = seeder.ToJSON()
			}
			if err != nil {
				logger.Error("Failed to render seed data: " + err.Error())
				os.Exit(1)
			}

			if out == "" {
				if format == "sql" {
					out = "seed.sql"
				} else {
					out = "seed.json"
				}
			}
			if err := os.WriteFile(out, []byte(result), 0o644); err != nil {
				logger.Error("Failed to write " + out + ": " + err.Error())
				os.Exit(1)
			}

			logger.Success(fmt.Sprintf("Wrote %s (%d record(s) per schema, seed=%d)", out, countOr(count), seed))
			logger.Print("")
			logger.Print("Load it:  psql < seed.sql   |   json: use as fixtures or `dev db seed`")
		},
	}

	cmd.Flags().IntVar(&count, "count", 5, "records per schema")
	cmd.Flags().Int64Var(&seed, "seed", 42, "PRNG seed (deterministic output)")
	cmd.Flags().StringVar(&format, "format", "json", "json | sql")
	cmd.Flags().StringVar(&out, "out", "", "output file (default seed.json / seed.sql)")
	return cmd
}

func countOr(c int) int {
	if c <= 0 {
		return 5
	}
	return c
}
