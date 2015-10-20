# Kubernetes Definitions

Kubernetes defines services, replication controllers, volumes, secrests
and other entities inside of _definitions_. They are encoded either as
JSON files or YAML files.

This directory is where those files should go.

Open questions:

- JSON, YAML, or don't care?
- Combine definitions into one file, or provide separate files, or don't
  care?
- Treat these as templates, or as release artifacts?
