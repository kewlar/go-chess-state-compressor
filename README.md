# Go Chess State Compressor

This is an experiment in packing chess game state into the smallest amount of data possible.

I saw [the question](https://codegolf.stackexchange.com/questions/19397/smallest-chess-board-compression), thought it to be an interesting challenge, got carried away, and this collection of codecs is the result.

## The rules

A chess state codec must:

* Encode positions possible in standard chess.
* Encode all the information a standard [FEN](https://en.wikipedia.org/wiki/Forsyth%E2%80%93Edwards_Notation) encodes.
* Assume halfmove clock may go up to 150: the biggest number of halfmoves possible until the seventy-five-move rule kicks in.
* Assume move number may go up to 8849: the [maximum possible number of moves in a chess game](https://www.wismuth.com/chess/longest-game.html).
* Assume compressed data is stored in memory. Include data length or end-of-data marker in the data if the end of data cannot be determined from the data otherwise.

It is not necessary to encode impossible (ex. more than 18 Queens, or more than 16 pieces on one side etc.) or non-standard positions (ex. Chess960 positions), but it's also not necessary to validate against it.

## How to run

Compile the package with `make`, and then run `bin/test_codecs` on a FEN file. Example:

```bash
bin/test_codecs data/common.txt
```

## Compression performance

When run on `comprehensive.txt` dataset, the codecs perform as follows:

| Codec                              |    Min, b |   Mean, b | Median, b |    Max, b | Worst case FEN |
| :--------------------------------- | --------: | --------: | --------: | --------: | :------------- |
| FEN                                |       248 | 492.67993 |       512 |       672 | b1q1k1r1/r1n1n1b1/1p1p1p1p/p1p1p1p1/P1P1P1P1/1P1P1P1P/R1N1N1B1/B1Q1K1R1 w - - 18 18 |
| BinaryFen                          |       106 | 194.50699 |       198 |       226 | 1b1q3r/p4r1p/1p2n1pP/kbpQ1p2/1NnpB1PR/1PP1p3/P2PPPK1/N1B1R3 b - - 114 1632 |
| Huffman                            |       108 | 181.82919 |       192 |       214 | 3qQq1q/1rkb1Q1q/Q6q/7q/Q7/Q5K1/Q1q3B1/QqQq1QR1 b - - 0 80 |
| Huffman (adaptive)                 |       100 | 169.47539 |       177 |       207 | 8/1rkbqQ1q/Q3Qq1q/1qq4q/Q7/Q4QK1/QQqQq1BR/8 b - - 12 86 |
| Arithmetic                         |       114 | 188.07213 |       198 |       220 | 3qQq1q/1rkb1Q1q/Q6q/7q/Q7/Q5K1/Q1q3B1/QqQq1QR1 b - - 0 80 |
| Arithmetic (by piece count)        |       107 | 188.00692 |       197 |       219 | 3qQq1q/1rkb1Q1q/Q6q/7q/Q7/Q5K1/Q1q3B1/QqQq1QR1 b - - 0 80 |
| Arithmetic (adaptive)              |       107 | 176.52605 |       184 |       225 | 8/1rkbqQ1q/Q3Qq1q/1qq4q/Q7/Q4QK1/QQqQq1BR/8 b - - 12 86 |
| BinaryFen (compact)                |        94 | 182.50699 |       186 |       214 | 1b1q3r/p4r1p/1p2n1pP/kbpQ1p2/1NnpB1PR/1PP1p3/P2PPPK1/N1B1R3 b - - 114 1632 |
| Huffman (compact)                  |       100 | 173.82919 |       184 |       206 | 3qQq1q/1rkb1Q1q/Q6q/7q/Q7/Q5K1/Q1q3B1/QqQq1QR1 b - - 0 80 |
| Huffman (adaptive, compact)        |        21 | 157.70968 |       166 |       195 | 8/1rkbqQ1q/Q3Qq1q/1qq4q/Q7/Q4QK1/QQqQq1BR/8 b - - 12 86 |
| Arithmetic (adaptive, compact)     |        21 | 164.74117 |       173 |       213 | 8/1rkbqQ1q/Q3Qq1q/1qq4q/Q7/Q4QK1/QQqQq1BR/8 b - - 12 86 |
| Arithmetic (adaptive, compact, v2) |        21 | 164.01722 |       173 |       214 | 8/1rkbqQ1q/Q3Qq1q/1qq4q/Q7/Q4QK1/QQqQq1BR/8 b - - 12 86 |
| Huffman (adaptive, compact, v2)    |        21 | 156.98573 |       166 |       196 | 8/1rkbqQ1q/Q3Qq1q/1qq4q/Q7/Q4QK1/QQqQq1BR/8 b - - 12 86 |

The `comprehensive.txt` dataset includes synthetic edge cases and eliminates duplicates of common positions, thus the compression statistics do not quite match the statistics on datasets from real games, but you can see the trends.

## Test data

There is some test data in [`data/`](data/):

| File              |  FENs | Notes                                                                                                             |
| :---------------- | ----: | :---------------------------------------------------------------------------------------------------------------- |
| real.txt          | 10000 | A representative sample from real world games from Lichess database.                                              |
| common.txt        |  1000 | The most frequently encountered positions in real-world games (mostly openings). Extracted from Lichess database. |
| synthetic.txt     |    12 | Artificial (but possible) positions intended to test how the codecs perform with edge cases.                      |
| longest_game.txt  | 17698 | Positions in the longest possible chess game.                                                                     |
| comprehensive.txt | 27756 | The combination of the above files, with duplicates removed.                                                      |

The format is plain text, 1 FEN per line. A non-standard extension is that FEN lines can have comments beginning with `#`, as seen in `synthetic.txt`.

More FEN data can be created from PGN files, using `pgn2fen`. For example:

```bash
bin/pgn2fen my_games.pgn > data/my_games.txt
```

Or pipe a PGN stream from an archive:

```bash
unzstd -c lichess_db_standard_rated_2013-01.pgn.zst | bin/pgn2fen - > data/lichess_db_standard_rated_2013-01.txt
```

## Code organization

The codecs are in `internal/codecs/*.go`. Auxiliary code is in the rest of `internal/`, with the essential for any codec being chess state model in `internal/chess/` and bitfield model in `internal/bitfield/`.

If you'd like to create a codec yourself, `BinaryFen` is the simplest real codec you can check out to see how it all comes together.

## Contribution

Have a go at it. Chess state compression is an interesting little problem to ponder, and I encourage you to try your hand at it. If you find a potential improvement or an interesting approach, open a pull request.

## AI disclaimer

The `Makefile` was created by an LLM.

