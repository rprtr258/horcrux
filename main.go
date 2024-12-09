package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"iter"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func factorial(n int) int {
	res := 1
	for i := range n {
		res *= i + 1
	}
	return res
}

func choose(k, n int) int {
	res := 1 // n!/(n-k)!
	kf := 1  // k!
	for i := range k {
		j := i + 1
		res *= n - k + j
		kf *= j
	}
	return res / kf
}

func combine(k, n int) iter.Seq2[int, []int] {
	return func(yield func(int, []int) bool) {
		comb := make([]int, k)
		for i := 0; i < k; i++ {
			comb[i] = i
		}

		for l := 0; ; l++ {
			if !yield(l, comb) {
				return
			}
			i := k - 1
			for i >= 0 && comb[i] == i+n-k {
				i--
			}
			if i < 0 {
				break
			}
			comb[i]++
			for j := i + 1; j < k; j++ {
				comb[j] = comb[j-1] + 1
			}
		}
	}
}

type header struct {
	N      int
	K      int
	Chunks []int
	Data   [][]byte
	// TODO: chunks as index+buffers
}

func main() {
	app := cli.App{
		Name: "horcrux",
		Commands: []*cli.Command{
			{
				Name: "split",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:     "n",
						Required: true,
						Action: func(_ *cli.Context, n int) error {
							if n <= 0 {
								return cli.Exit("n must be greater than 0", 1)
							}
							return nil
						},
					},
					&cli.IntFlag{
						Name:     "k",
						Required: true,
						Action: func(_ *cli.Context, k int) error {
							if k <= 1 {
								return cli.Exit("k must be greater than 1", 1)
							}
							return nil
						},
					},
				},
				Args:      true,
				ArgsUsage: "<file>",
				Action: func(c *cli.Context) error {
					filename := c.Args().Get(0)
					n := c.Int("n")
					k := c.Int("k")
					{ // validation
						if k > n {
							return cli.Exit("k must be less than or equal to n", 1)
						}
						if filename == "" {
							return cli.Exit("filename is required", 1)
						}
						// TODO: check file exists and is not directory
						// TODO: check file is of size >= chunks
						// TODO: horcrux files do not exist
					}

					chunks := choose(k-1, n)
					if chunks > 1000 {
						return cli.Exit(enewf("too many chunks: %d", chunks), 1)
					}

					horcruxChunks := make([][]int, n)
					for chunk, is := range combine(k-1, n) {
						// chunk MUST NOT be in those horcruxes
						// AND must be in all others horcruxes
						j := 0
						for i := range n {
							if j < k-1 && is[j] == i {
								j++
								continue
							}

							horcruxChunks[i] = append(horcruxChunks[i], chunk)
						}
						chunk++
					}

					file, err := os.Open(filename)
					if err != nil {
						return cli.Exit(ewrap(err, "open file"), 1)
					}
					defer file.Close()

					{
						chunksBytes := make([][]byte, chunks)
						b := make([]byte, chunks)
						// TODO: first, read fully chunks bytes, so that each chunk is not empty
						// then, read upto chunks bytes, so that chunks are of equal length
						for {
							n, err := file.Read(b)
							for chunk, c := range b[:n] {
								chunksBytes[chunk] = append(chunksBytes[chunk], c)
							}
							if err != nil {
								if err == io.EOF {
									break
								}
								return cli.Exit(ewrap(err, "read pre-chunk"), 1)
							}
						}

						for i, chunks := range horcruxChunks {
							m := header{
								N:      n,
								K:      k,
								Chunks: chunks,
							}
							for _, chunk := range chunks {
								m.Data = append(m.Data, chunksBytes[chunk])
							}

							if err := func() error {
								horcruxFile, err := os.OpenFile(fmt.Sprintf("%s.%d.horcrux", filename, i), os.O_CREATE|os.O_WRONLY, 0644)
								if err != nil {
									return ewrapf(err, "create %d horcrux file", i)
								}
								defer horcruxFile.Close()

								if err := gob.NewEncoder(horcruxFile).Encode(m); err != nil {
									return ewrapf(err, "encode %d horcrux file", i)
								}

								return nil
							}(); err != nil {
								return cli.Exit(err, 1)
							}
						}
					}

					return nil
				},
			},
			{
				Name: "join",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "output",
						Aliases:  []string{"o"},
						Required: false,
						Value:    "output.txt",
						Usage:    "<file>",
					},
				},
				Action: func(c *cli.Context) error {
					outfile := c.String("output")
					if c.Args().Len() == 0 {
						return cli.Exit("filenames required", 1)
					}
					// TODO: check outfile does not exist
					// TODO: check input files exist

					headers := []header{}
					for _, filename := range c.Args().Slice() {
						if err := func() error {
							horcruxFile, err := os.Open(filename)
							if err != nil {
								return cli.Exit(ewrapf(err, "open %s", filename), 1)
							}
							defer horcruxFile.Close()

							var m header
							if err := gob.NewDecoder(horcruxFile).Decode(&m); err != nil {
								return cli.Exit(ewrapf(err, "decode %s", filename), 1)
							}
							headers = append(headers, m)

							return nil
						}(); err != nil {
							return cli.Exit(err, 1)
						}
					}

					n := headers[0].N
					k := headers[0].K
					if len(headers) < k {
						return cli.Exit(enewf("not enough horcruxes, at least %d are required", k), 1)
					}
					chunks := choose(k-1, n)
					chunksBytes := make([][]byte, chunks)
					chunksSeen := make([]bool, chunks)
					for _, h := range headers {
						if h.N != n {
							return cli.Exit("all headers must have same n", 1)
						}
						if h.K != k {
							return cli.Exit("all headers must have same k", 1)
						}
						for i, c := range h.Chunks {
							chunkB := h.Data[i]
							if !chunksSeen[c] {
								chunksBytes[c] = chunkB
								chunksSeen[c] = true
							} else {
								if !bytes.Equal(chunksBytes[c], chunkB) {
									return cli.Exit("inconsistent chunk", 1)
								}
							}
						}
					}
					// TODO: chunks lengths are non-increasing inside chunksBytes
					for i, seen := range chunksSeen {
						if !seen {
							return cli.Exit(enewf("missing chunk %d", i), 1)
						}
					}

					file, err := os.OpenFile(outfile, os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						return cli.Exit(ewrap(err, "open output file"), 1)
					}
					defer file.Close()

					i := 0
				LOOP:
					for {
						for _, b := range chunksBytes {
							if i >= len(b) {
								break LOOP
							}
							if _, err := file.Write([]byte{b[i]}); err != nil {
								return cli.Exit(ewrap(err, "write chunk"), 1)
							}
						}
						i++
					}

					return nil
				},
			},
		},
	}
	log.SetFlags(0)
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err.Error())
	}
}
