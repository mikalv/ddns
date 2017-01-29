FROM buildpack-deps:jessie-scm
MAINTAINER Philipp Böhm <philipp@pboehm.org>

ENV GIN_MODE release

CMD /go/bin/app --domain=d.pboehm.de --redis=redis:6379 web
