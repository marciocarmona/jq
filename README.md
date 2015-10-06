# jq Go binding.

This is a Go binding for Jq lib based on github.com/mattaitchison/jq POC. It also includes a HTTP handler for Jq expression application on an JSON response. 

Many memory use improvements were made and it also uses a LRU cache for Jq expressions as compiling them is usually more expensive than executing.

Some benchmark tests are available to check performance.

You need to have Jq installed to use it.