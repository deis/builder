#!/bin/bash
set -e

fetcher &
exec builder
