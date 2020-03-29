package(default_visibility = ["PUBLIC"])

go_get(
    name = "logging",
    get = "gopkg.in/op/go-logging.v1",
    revision = "b2cb9fa56473e98db8caba80237377e83fe44db5",
)

go_get(
    name = "terminal",
    get = "golang.org/x/crypto/ssh/terminal",
    revision = "7b85b097bf7527677d54d3220065e966a0e3b613",
)

go_get(
    name = "go-bindata",
    binary = True,
    get = "github.com/kevinburke/go-bindata/...",
    revision = "46eb4c183bfc1ebb527d9d19bcded39476302eb8",
    strip = ["testdata"],
)

go_get(
    name = "net",
    get = "golang.org/x/net/...",
    revision = "136a25c244d3019482a795d728110278d6ba09a4",
    deps = [
        ":terminal",
        ":text",
    ],
)

go_get(
    name = "tools",
    get = "golang.org/x/tools",
    install = [
        "cover",
    ],
    revision = "2ae76fd1560b622911f444c1e66b70a857e1de67",
)

go_get(
    name = "text",
    get = "golang.org/x/text/...",
    revision = "4e4a3210bb54bb31f6ab2cdca2edcc0b50c420c1",
    deps = [
        ":tools",
    ],
)

go_get(
    name = "protobuf",
    get = "github.com/golang/protobuf/...",
    revision = "v1.3.2",
    strip = [
        "proto/proto3_proto",
        "conformance",
    ],
)

go_get(
    name = "errgroup",
    get = "golang.org/x/sync/errgroup",
    revision = "457c5828408160d6a47e17645169cf8fa20218c4",
    deps = [":net"],
)

go_get(
    name = "skylark",
    get = "github.com/google/skylark/syntax",
    revision = "a5f7082aabed29c0e429c722292c66ec8ecf9591",
)

go_get(
    name = "buildtools",
    get = "github.com/bazelbuild/buildtools/...",
    revision = "0.19.2",
    deps = [
        ":protobuf",
        ":skylark",
    ],
)

go_get(
    name = "buildifier",
    binary = True,
    get = [],
    install = ["github.com/bazelbuild/buildtools/buildifier"],
    deps = [
        ":buildtools",
    ],
)

go_get(
    name = "mux",
    get = "github.com/gorilla/mux",
    revision = "v1.7.4",
)
