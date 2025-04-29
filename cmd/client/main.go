package main

import (
	"context"
	"log"

	"github.com/mdubakin/balun-grpc-course/pkg/api/example"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	conn, err := grpc.NewClient(":8085", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	client := example.NewExampleClient(conn)

	cctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("client-header-key", "val"))
	req := &example.CreatePostRequest{
		Title:    "example",
		AuthorId: "1",
		Content:  "hello",
	}
	var header, trailer metadata.MD
	resp, err := client.CreatePost(
		cctx,
		req,
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)
	if err != nil {
		switch status.Code(err) {
		case codes.InvalidArgument:
			log.Println("некорректный запрос")
		default:
			log.Fatal(err)
		}

		if st, ok := status.FromError(err); ok {
			log.Println("code", st.Code(), "message", st.Message(), "details", st.Details())
		} else {
			log.Println("not grpc")
		}
	}
	log.Println("headers: ", header, "trailers: ", trailer)
	log.Println(resp.GetPostId())

	// пакет protojson для де/сериализации JSON из proto сообщений
	bytes, err := protojson.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(bytes))
}
