#!/bin/bash
set -e

bpbuilder & 
exec fetcher
