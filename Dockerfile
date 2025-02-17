# Copyright 2022 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:alpine@sha256:fd78f2fb1e49bcf343079bbbb851c936a18fc694df993cbddaa24ace0cc724c5 AS builder

WORKDIR /src
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./
RUN go build -o osv-scanner ./cmd/osv-scanner/

FROM alpine:3.19@sha256:51b67269f354137895d43f3b3d810bfacd3945438e94dc5ac55fdac340352f48

RUN apk --no-cache add ca-certificates git && \
    git config --global --add safe.directory '*'

WORKDIR /root/
COPY --from=builder /src/osv-scanner .

ENTRYPOINT ["/root/osv-scanner"]
