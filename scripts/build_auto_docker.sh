#!/bin/bash
VERSION=5.0
docker build -t oes-acc-test:$VERSION -f scripts/Dockerfile .