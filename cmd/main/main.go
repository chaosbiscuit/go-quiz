package main

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type Args struct {
	csv   string
	limit int
}

func ParseArgs() Args {
	args := Args{}
	flag.IntVar(&args.limit, "limit", 30, "Quiz time limit in seconds")
	flag.StringVar(&args.csv, "csv", "questions.csv", "Path to questions .csv")
	flag.Parse()
	return args
}

func AskQuestion(ctx context.Context, q string, a string) (bool, error) {
	fmt.Println(q)

	ch := make(chan string)
	var input string
	go func() {
		fmt.Scanln(&input)
		ch <- input
	}()

	// Block until question is answered or cancellation occurs
	select {
	case <-ctx.Done():
		fmt.Println("Time's up")
		return false, ctx.Err()
	case <-ch:
	}

	return input == a, nil
}

func Quizzer(ctx context.Context, records [][]string) {
	correct := 0
	answered := 0
	for _, question := range records {
		res, err := AskQuestion(ctx, question[0], question[1])
		if err != nil {
			break
		}

		if res {
			correct++
		}
		answered++
	}
	fmt.Printf("Answered %v/%v correctly - Score %.2f%%\n", correct, len(records), (float32(correct) / float32(len(records)) * 100.00))
}

func main() {

	args := ParseArgs()

	fd, err := os.Open(args.csv)
	if err != nil {
		panic(errors.Join(fmt.Errorf("unable to open problems .csv"), err))
	}

	reader := csv.NewReader(fd)
	questions, err := reader.ReadAll()
	if err != nil {
		panic(errors.Join(fmt.Errorf("nvalid .csv formatting"), err))
	}
	rand.Shuffle(len(questions), func(i, j int) {
		questions[i], questions[j] = questions[j], questions[i]
	})

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(args.limit)*time.Second)
	defer cancelFunc()
	Quizzer(ctx, questions)
}
