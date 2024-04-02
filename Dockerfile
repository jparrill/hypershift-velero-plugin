# Copyright 2024 Red Hat Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM quay.io/konveyor/builder as builder
ENV GOPROXY=https://proxy.golang.org
ENV GOPATH=$APP_ROOT
WORKDIR $APP_ROOT/src/github.com/jparrill/hypershift-velero-plugin
COPY go.mod go.sum $APP_ROOT/src/github.com/jparrill/hypershift-velero-plugin/
RUN go mod download
COPY . $APP_ROOT/src/github.com/jparrill/hypershift-velero-plugin
RUN CGO_ENABLED=0 GOARCH=amd64 go build -o /go/bin/hypershift-velero-plugin .

FROM registry.access.redhat.com/ubi8-minimal AS ubi8
COPY --from=builder /go/bin/hypershift-velero-plugin /plugins/
USER 65532:65532
ENTRYPOINT ["/bin/bash", "-c", "cp /plugins/hypershift-velero-plugin /target/."]