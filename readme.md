# Nomad-Listener

> Listen to the Nomad event stream, and run handlers

## Usage

`nomad-listener` will listen to the Nomad event stream, and for each event, check if the `handlers` directory contains an executable for the given topic and event combination.

Handlers are named in the form `$topic-$event`, for example `Job-JobRegistered`, and must be in the `handlers` directory:

```
├── handlers
│   └── Job-JobRegistered
└── nomad-listener
```

