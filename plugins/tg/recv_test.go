package tg_test

import (
	"context"
	"io"
	"math/big"
	"testing"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mock"
	"github.com/redjack/marionette/plugins/tg"
	"go.uber.org/zap"
)

func init() {
	if !testing.Verbose() {
		marionette.Logger = zap.NewNop()
	}
}

func TestRecv(t *testing.T) {
	t.Skip("TODO: Replace mock dfa with actual dfa")

	t.Run("OK", func(t *testing.T) {
		var stream *marionette.Stream
		streamSet := marionette.NewStreamSet()
		streamSet.OnNewStream = func(s *marionette.Stream) { stream = s }

		conn := mock.DefaultConn()
		conn.ReadFn = func(p []byte) (int, error) {
			ret := []byte("GET http://127.0.0.1:8080/&&&&tbrAaIr.JQYjwzTUpVg7vI-cEmiOj7IyJjVVqR2Lra3McF3w-L6G5lgCmvkMQ11748uxAgFe-.3M2NVZPnbe2GyFk84LmxJZTLmUR0c67HHpqLX.mLGpLRpUX70GJIHrl29i?ogPU8O2gNngk3vBkHXWF1oykGk2C002T.t2j7IanHHj0QfKn9ru2qD??xENmrr81NtAYPdXOMs3jV6uVh3FjUlyNsF0vc7sC7VrclOJ4pXiXxl-qjYopqEtLKpNi.cXvE3B6X6YZ0j.PWtktDlGPHz9NhXdo6N2ZFx2Yw4VDSDwrLCq5YD-MZ0KA&AxOcf6jCiDMFlR0oXvQjTinDlw2dzaG2.BwVRFoC7jYuAY?uNZ3l2CewU6anfZwwjtMTB1j8-yERP5rC3A1EAPhYuro50Dl3k.siTnVejRyrPis?0?MCj4eGLJj&WLNgkE8Pzhoqc0Pv5ynYXzQajxY8ZRUBbqBNPf1JunpCIQc1w9wUqIM3FttgK65c3ztpjTRi24NLxAzGQSZcrAbX0iPu1l&xcoB&9IpDpXDydHYzfuDF53v0&iJpS86f3pzkQpRgB5jvhAY2a9HdhE1jsezlWNU6WvWfVZp2lLELpxpk?-YBa4R9Re4tZv-lS15hxwLQgwVg4B67Qc4Ojz9YhUNmSjap3xwIBw3BO3nukhB8xSYAe4q1s1vhJJYALcO1boahszGOk0BMZg&GpDpeEwT?9roz4?ymDs3fC?JUgtVVZGZ2-75pski7U7MiasrVDaJ9G9TvR?3qfIu-yij8JMbppSRPSz&corWUNhfjrefZhaJxqUMO1bcxu6eQIbdW-f2VdLI4nDZtDtPRCLzG.7HswyvvvAdM7a4CwcBh7sa9hU4Ryqsa8ZP0Fw8xQMsC4TrIB9KLD5LefWsa0mtE47VZSv?RB&bRtVWod2IoruqmKz98jWR5MeIFo6TNm?gMlYs6x&7YTl7iG1bwFCsfNsanJ3Iycm3dnOtS7ZVgOxTlmt5ohHYGam2Q.r&Lw8G8W77snmZc9L3ooES&TPmd3g9XI6?dCmLNJg9ETIWNKKszaGkIpF7udKYvUtlIAcoyu5CGdw?dfUu6mPLTgkBWyvMOd7.PomM8NNmMaweA8zfhpDtS2ZDAp?XH22fRr-ImatHiLK-KsLAEmeMv-H2U&wZFSn3BG2PCxG-HkZGCA5c.BeQlYx&DT9DBA381.fD4dm-isJWJXrG.v-i50s7o00I-SVkZ6&b31ZSu5wK3Sjnaq?W&Ex.GytkYHj.dbN9puXGi5ROglFHr8MQr1nYKSK29pIJoa8HE?-lbji2aJsw2aCJTG?ec3bSF1spg&1rOjUNLY77IRj.85tB.CW5M?dDj7egIe.SCVdz?x3T5hMZMH7eQyGixfs&35s3Q&hnD2l0F.7b&4DPEkh183zlaWe0imKJqqRcI1ruigG2Zb0oLI5gMoh-dngwJ7DzRRsC7CCnSw8hRcdIxYz1opATcGB3gQr-w9uIdSKi&BtUfvH0?Sjmauq6CCL9aV9Cy.MtpCfIU7t9AgSbfH3?OeVFg1OjfzOu-VVEeY0OtkvtqBguLtZWB--TZ49YqzNL72KVkqZlVtMs1iEmqzCZfGzRgzQOKd.5kvLCa1zhn-qIPem0RrcqpQQt6ogmnLVbgSBPshktGrAKDODIPS3?V?cLkOh1dAIdhHtL9eo9GWttHFEgpmBylubVdvWuKP9tYMlydzbnoItm3iYTY6EKLWuTH3Kaagalq6o8i6bqoKwH.4LJkc8hZyAhnVW2HYq13736s6qWlYLb8AuGv39vu47nyq0Ji1mM?z31d5VzeNmgKbuyRSuoghPmy5zrBDhh3IeJeEXRbRS3vhlbHy3XyXymQ02G.TGRmSXal2fIUf4s?ovKrpfxCY?iI16hkyrTm.GbEEz1pkb3Qwnq1EzdO0t&-63iAtKjhO5B7en46PcHPPQR9jABX3q&Y6?NEoMBqJhDZ6Pa4ZfrGzg0yifbqAOl-Klz9ai4UuQ9RYx01H4UOR5S?hcXtIu3Vv16OSFgw8E HTTP/1.1\r\nUser-Agent: marionette 0.1\r\nConnection: close\r\n\r\n")
			copy(p, ret)
			return len(ret), nil
		}

		var dfa mock.DFA
		dfa.CapacityFn = func() int { return 1000 }
		dfa.RankFn = func(s string) (rank *big.Int, err error) {
			return big.NewInt(123), nil
		}
		fsm := mock.NewFSM(&conn, streamSet)
		fsm.PartyFn = func() string { return marionette.PartyClient }
		fsm.HostFn = func() string { return "127.0.0.1" }
		fsm.UUIDFn = func() int { return 100 }
		fsm.InstanceIDFn = func() int { return 200 }
		fsm.DFAFn = func(regex string, msgLen int) (marionette.DFA, error) {
			return &dfa, nil
		}

		if err := tg.Recv(context.Background(), &fsm, `http_request_close`); err != nil {
			t.Fatal(err)
		} else if stream == nil {
			t.Fatal("expected stream")
		}

		// Read from stream.
		buf := make([]byte, 3)
		if _, err := io.ReadFull(stream, buf); err != nil {
			t.Fatal(err)
		} else if string(buf) != `foo` {
			t.Fatalf("unexpected read: %q", buf)
		}
	})

	t.Run("ErrNotEnoughArguments", func(t *testing.T) {
		fsm := mock.NewFSM(&mock.Conn{}, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		if err := tg.Recv(context.Background(), &fsm); err == nil || err.Error() != `tg.recv: not enough arguments` {
			t.Fatalf("unexpected error: %q", err)
		}
	})

	t.Run("ErrInvalidArgument", func(t *testing.T) {
		fsm := mock.NewFSM(&mock.Conn{}, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		if err := tg.Recv(context.Background(), &fsm, 123); err == nil || err.Error() != `tg.recv: invalid grammar name argument type` {
			t.Fatalf("unexpected error: %q", err)
		}
	})

	t.Run("ErrGrammarNotFound", func(t *testing.T) {
		fsm := mock.NewFSM(&mock.Conn{}, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		if err := tg.Recv(context.Background(), &fsm, "no_such_grammar"); err == nil || err.Error() != `tg.recv: grammar not found` {
			t.Fatalf("unexpected error: %q", err)
		}
	})
}
