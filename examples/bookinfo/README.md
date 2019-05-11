

# Book info sample app 

This app has been adapted from [Istio's sample app](https://github.com/istio/istio/tree/master/samples/bookinfo)


# Summary of changes

Some of the images in `bookinfo.yaml` have been adapted for the testing needs of Supergloo.

Changes are summarized below.

## reviews v4: Propagation of failure in review service

- purpose:
   - example of a non-robust service: it produces 500's when one of the services that it interacts with produces a bad response.
- description:
  - failures in requests from the review service to the rating service produce failures (`500` responses) in requests from the product page to the review page

- diagram:
  - `product` <-a-> `reviews:v4` <-b-> `ratings`
  - condition: no faults
    - description: weak point is "hidded" (not expressed)
    - result: reviews service behaves in same manner as `reviews:v3`
  - condition: fault in route `b`
    - description: weak point is expressed, cascading failure results
    - result: error is propagated to route `a`
    - preferred behavior: the `reviews` service should be able to provide a valid response even if it encounters errors from the `ratings` service

empty response from
