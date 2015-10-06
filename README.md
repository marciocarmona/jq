# jq Go binding.

This is a Go binding for Jq lib based on github.com/mattaitchison/jq POC. It also includes a HTTP handler for Jq expression apply over an JSON response. 

Many memory improvements had been made and it uses a LRU cache for Jq expressions as compiling them is usually more expensive than executing. 