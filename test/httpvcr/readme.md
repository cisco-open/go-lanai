# HTTP VCR

## Original Source
The original source code for httpvcr comes from [go-vcr](https://github.com/dnaeon/go-vcr)

## Why it was Copied
We needed something similar to copyist but for HTTP interactions. `go-vcr` hit all
the checkboxes except for requiring responses to be in order. To add that functionality, we needed
to add or extend the `Cassette` struct. However, the `Recorder` struct doesn't export any of 
its fields so we needed to directly edit `Cassette`.

## Version History
The version that was originall copied over was: [v2.0.1](https://github.com/dnaeon/go-vcr/releases/tag/v2.0.1)