package fte_test

import (
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/redjack/marionette/fte"
)

func TestDFA(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		dfa, err := fte.NewDFA(`[a-zA-Z0-9\?\-\.\&]+`, 2048)
		if err != nil {
			t.Fatal(err)
		}
		defer dfa.Close()

		if capacity := dfa.Capacity(); capacity != 1547 {
			t.Fatalf("unexpected capacity: %d", capacity)
		}

		msg0 := strings.Repeat("A", 2048)
		msg1 := strings.Repeat("B", 2048)

		// Encode/decode first message.
		if rank, err := dfa.Rank(msg0); err != nil {
			t.Fatal(err)
		} else if other, err := dfa.Unrank(rank); err != nil {
			t.Fatal(err)
		} else if other != msg0 {
			t.Fatalf("unexpected unrank: %q", other)
		}

		// Encode/decode second message.
		if rank, err := dfa.Rank(msg1); err != nil {
			t.Fatal(err)
		} else if other, err := dfa.Unrank(rank); err != nil {
			t.Fatal(err)
		} else if other != msg1 {
			t.Fatalf("unexpected unrank: %q", other)
		}

		if err := dfa.Close(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("TG", func(t *testing.T) {
		dfa, err := fte.NewDFA(`[a-zA-Z0-9\?\-\.\&]+`, 2048)
		if err != nil {
			t.Fatal(err)
		}
		defer dfa.Close()

		msg := `&&&&tbrAaIr.JQYjwzTUpVg7vI-cEmiOj7IyJjVVqR2Lra3McF3w-L6G5lgCmvkMQ11748uxAgFe-.3M2NVZPnbe2GyFk84LmxJZTLmUR0c67HHpqLX.mLGpLRpUX70GJIHrl29i?ogPU8O2gNngk3vBkHXWF1oykGk2C002T.t2j7IanHHj0QfKn9ru2qD??xENmrr81NtAYPdXOMs3jV6uVh3FjUlyNsF0vc7sC7VrclOJ4pXiXxl-qjYopqEtLKpNi.cXvE3B6X6YZ0j.PWtktDlGPHz9NhXdo6N2ZFx2Yw4VDSDwrLCq5YD-MZ0KA&AxOcf6jCiDMFlR0oXvQjTinDlw2dzaG2.BwVRFoC7jYuAY?uNZ3l2CewU6anfZwwjtMTB1j8-yERP5rC3A1EAPhYuro50Dl3k.siTnVejRyrPis?0?MCj4eGLJj&WLNgkE8Pzhoqc0Pv5ynYXzQajxY8ZRUBbqBNPf1JunpCIQc1w9wUqIM3FttgK65c3ztpjTRi24NLxAzGQSZcrAbX0iPu1l&xcoB&9IpDpXDydHYzfuDF53v0&iJpS86f3pzkQpRgB5jvhAY2a9HdhE1jsezlWNU6WvWfVZp2lLELpxpk?-YBa4R9Re4tZv-lS15hxwLQgwVg4B67Qc4Ojz9YhUNmSjap3xwIBw3BO3nukhB8xSYAe4q1s1vhJJYALcO1boahszGOk0BMZg&GpDpeEwT?9roz4?ymDs3fC?JUgtVVZGZ2-75pski7U7MiasrVDaJ9G9TvR?3qfIu-yij8JMbppSRPSz&corWUNhfjrefZhaJxqUMO1bcxu6eQIbdW-f2VdLI4nDZtDtPRCLzG.7HswyvvvAdM7a4CwcBh7sa9hU4Ryqsa8ZP0Fw8xQMsC4TrIB9KLD5LefWsa0mtE47VZSv?RB&bRtVWod2IoruqmKz98jWR5MeIFo6TNm?gMlYs6x&7YTl7iG1bwFCsfNsanJ3Iycm3dnOtS7ZVgOxTlmt5ohHYGam2Q.r&Lw8G8W77snmZc9L3ooES&TPmd3g9XI6?dCmLNJg9ETIWNKKszaGkIpF7udKYvUtlIAcoyu5CGdw?dfUu6mPLTgkBWyvMOd7.PomM8NNmMaweA8zfhpDtS2ZDAp?XH22fRr-ImatHiLK-KsLAEmeMv-H2U&wZFSn3BG2PCxG-HkZGCA5c.BeQlYx&DT9DBA381.fD4dm-isJWJXrG.v-i50s7o00I-SVkZ6&b31ZSu5wK3Sjnaq?W&Ex.GytkYHj.dbN9puXGi5ROglFHr8MQr1nYKSK29pIJoa8HE?-lbji2aJsw2aCJTG?ec3bSF1spg&1rOjUNLY77IRj.85tB.CW5M?dDj7egIe.SCVdz?x3T5hMZMH7eQyGixfs&35s3Q&hnD2l0F.7b&4DPEkh183zlaWe0imKJqqRcI1ruigG2Zb0oLI5gMoh-dngwJ7DzRRsC7CCnSw8hRcdIxYz1opATcGB3gQr-w9uIdSKi&BtUfvH0?Sjmauq6CCL9aV9Cy.MtpCfIU7t9AgSbfH3?OeVFg1OjfzOu-VVEeY0OtkvtqBguLtZWB--TZ49YqzNL72KVkqZlVtMs1iEmqzCZfGzRgzQOKd.5kvLCa1zhn-qIPem0RrcqpQQt6ogmnLVbgSBPshktGrAKDODIPS3?V?cLkOh1dAIdhHtL9eo9GWttHFEgpmBylubVdvWuKP9tYMlydzbnoItm3iYTY6EKLWuTH3Kaagalq6o8i6bqoKwH.4LJkc8hZyAhnVW2HYq13736s6qWlYLb8AuGv39vu47nyq0Ji1mM?z31d5VzeNmgKbuyRSuoghPmy5zrBDhh3IeJeEXRbRS3vhlbHy3XyXymQ02G.TGRmSXal2fIUf4s?ovKrpfxCY?iI16hkyrTm.GbEEz1pkb3Qwnq1EzdO0t&-63iAtKjhO5B7en46PcHPPQR9jABX3q&Y6?NEoMBqJhDZ6Pa4ZfrGzg0yifbqAOl-Klz9ai4UuQ9RYx01H4UOR5S?hcXtIu3Vv16OSFgw8E`
		if rank, err := dfa.Rank(msg); err != nil {
			t.Fatal(err)
		} else if other, err := dfa.Unrank(rank); err != nil {
			t.Fatal(err)
		} else if other != msg {
			t.Fatalf("unexpected unrank: %q", other)
		}
	})
}

func TestDFA_NumWordsInSlice(t *testing.T) {
	dfa, err := fte.NewDFA(`[a-zA-Z0-9\?\-\.\&]+`, 2048)
	if err != nil {
		t.Fatal(err)
	}
	if n, err := dfa.NumWordsInSlice(2); err != nil {
		t.Fatal(err)
	} else if n.Int64() != 4356 {
		t.Fatalf("unexpected num: %s", n.String())
	}
}

func TestLog2(t *testing.T) {
	for _, tt := range []struct {
		value  *big.Int
		result int
	}{
		{big.NewInt(1), 0},
		{big.NewInt(2), 1},
		{big.NewInt(3), 1},
		{big.NewInt(4), 2},
		{big.NewInt(7), 2},
		{big.NewInt(8), 3},
		{big.NewInt(math.MaxInt64), 62},
	} {
		t.Run(tt.value.String(), func(t *testing.T) {
			if diff := cmp.Diff(tt.result, fte.Log2(tt.value)); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
