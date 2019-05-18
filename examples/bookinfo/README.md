

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


### Initial setup
- note that the reviews and ratings are shown
```bash
supergloo init
kubectl label namespace default istio-injection=enabled
supergloo install istio --name istio --installation-namespace istio-system --mtls=true --auto-inject=true
kubectl apply -f bookinfo.yaml
supergloo apply routingrule trafficshifting \
    --name reviews-v4 \
    --dest-upstreams supergloo-system.default-reviews-9080 \
    --target-mesh supergloo-system.istio \
    --destination supergloo-system.default-reviews-v4-9080:1
kubectl port-forward -n default deployment/productpage-v1 9080
## OPEN localhost:9080 in your browser - see the "baseline" app
```

### What does it look like when `reviews` fail
- note that neither reviews nor ratings are shown
```bash
supergloo apply routingrule faultinjection abort http \
    --target-mesh supergloo-system.istio \
     -p 100 -s 500  --name fault-product-to-reviews \
    --dest-upstreams supergloo-system.default-reviews-9080
## RELOAD PAGE - Expect error from reviews
# cleanup:
kubectl delete routingrule -n supergloo-system fault-product-to-reviews
## RELOAD PAGE - Expect errors to have gone away
```

### Trigger the weakness
- note, again, that neither reviews nor ratings are shown
```bash
supergloo apply routingrule faultinjection abort http \
    --target-mesh supergloo-system.istio \
     -p 100 -s 500  --name fault-reviews-to-ratings \
    --dest-upstreams supergloo-system.default-ratings-9080
## RELOAD PAGE - Expect reviews to fail in the same manner
```

### Replace the weak service with a more robust version
- note that reviews show without ratings
```bash
kubectl delete routingrule -n supergloo-system reviews-v4
supergloo apply routingrule trafficshifting \
    --name reviews-v3 \
    --dest-upstreams supergloo-system.default-reviews-9080 \
    --target-mesh supergloo-system.istio \
    --destination supergloo-system.default-reviews-v3-9080:1
## RELOAD PAGE - Expect failure to be more graceful, reviews information is shown without the ratings
```

