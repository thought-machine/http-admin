subinclude("//build_defs:go_bindata")

go_library(
    name = "http-admin",
    srcs = glob(["*.go"], exclude = ["bindata.go"]) + [":bindata"],
    visibility = ["PUBLIC"],
    deps = [
        "//third_party/go:logging",
        "//third_party/go:mux",
        "//third_party/go:net",
        "//third_party/go/prometheus",
        "//third_party/go/prometheus:client_model",
    ],
)

go_bindata(
    name = "bindata",
    srcs = [
        "//css",
        "//img",
        "//js",
    ],
    package = "admin",
    all_dirs = True,
)
