package main

import (
	"context"
	"io"
	"log"
	"math/rand"
	"net"
	"sync"

	"github.com/bufbuild/protovalidate-go"
	"github.com/mdubakin/balun-grpc-course/pkg/api/example"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type Post struct {
	ID       uint64
	Title    string
	Content  string
	AuthorID string
}

type ExampleService struct {
	example.UnimplementedExampleServer

	validator protovalidate.Validator
	storage   map[uint64]*Post
	mx        sync.RWMutex
}

func (s *ExampleService) CreatePost(ctx context.Context, req *example.CreatePostRequest) (*example.CreatePostResponse, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		log.Println(md)
	}

	if err := s.validator.Validate(req); err != nil {
		st := status.New(codes.InvalidArgument, codes.InvalidArgument.String())
		st, _ = st.WithDetails(&errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "request",
					Description: err.Error(),
				},
			},
		})
		return nil, st.Err()
	}

	id := rand.Uint64()
	post := &Post{
		ID:       id,
		AuthorID: req.GetAuthorId(),
		Content:  req.GetContent(),
		Title:    req.GetTitle(),
	}
	s.mx.Lock()
	s.storage[id] = post
	s.mx.Unlock()

	header := metadata.Pairs("header-key", "val")
	grpc.SetHeader(ctx, header)
	grpc.SetTrailer(ctx, header)

	return &example.CreatePostResponse{
		PostId: id,
	}, nil

}
func (s *ExampleService) ListPosts(ctx context.Context, req *example.ListPostsRequest) (*example.ListPostsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListPosts not implemented")
}

func main() {
	server := grpc.NewServer()

	validator, err := protovalidate.New()
	if err != nil {
		log.Fatal(err)
	}

	service := &ExampleService{
		validator: validator,
		storage:   map[uint64]*Post{},
	}

	example.RegisterExampleServer(server, service)

	lis, err := net.Listen("tcp", ":8085")
	if err != nil {
		log.Fatal(err)
	}

	reflection.Register(server)

	log.Println("gRPC server listen on :8085")
	if err := server.Serve(lis); err != io.EOF {
		log.Fatal(err)
	}
}
