---
title: Bisection Bandwidth
category: networking
tags: [bisection, bandwidth, clos, throughput, non-blocking]
---

# Bisection Bandwidth

## What is bisection bandwidth

Bisection bandwidth is the total bandwidth across the narrowest cut that divides a network into two equal halves. It is the definitive measure of a fabric's ability to move traffic between any two servers simultaneously — regardless of where those servers sit in the topology.

To compute bisection bandwidth, imagine slicing the network topology graph in half (by server count) and summing the bandwidth of every link that crosses the cut. If you always find the same minimum, that minimum is the bisection bandwidth. A higher bisection bandwidth means that, even in the adversarial worst case, servers can communicate with more total bandwidth.

Bisection bandwidth is expressed in Gbps or Tbps for large fabrics. In fabrik, the bisection bandwidth shown in the metrics dashboard sums the uplink bandwidth of all leaf switches across all fabrics in the selected design.

## Full bisection

A full bisection fabric is one where the bisection bandwidth equals the total server-facing (downlink) bandwidth. In other words, every server in the network can simultaneously transmit to every other server at line rate with no congestion anywhere in the fabric. This is also called a non-blocking fabric.

Full bisection is the gold standard for latency-sensitive and bandwidth-intensive workloads: AI training (where GPUs need to exchange gradients at full speed), HPC clusters, and large distributed storage systems. In a leaf-spine Clos, full bisection requires that every leaf connects to every spine, and each spine has enough ports to terminate all leaves — typically requiring a 1:1 ratio between leaf uplink bandwidth and leaf downlink bandwidth.

In practice, full bisection fabrics are significantly more expensive than oversubscribed fabrics of the same server count, because the spine tier scales linearly with the number of leaf switches. Most hyperscale operators accept some oversubscription and use traffic engineering or application-level throttling to prevent hotspots.

## Oversubscription and bisection

Oversubscription directly reduces bisection bandwidth. An oversubscription ratio of 3:1 means the fabric can only deliver one-third of its theoretical full bisection bandwidth across the spine layer before congestion occurs. Stated differently, the bisection bandwidth of an oversubscribed fabric is:

```
Bisection BW = Total downlink BW / Oversubscription ratio
```

For example, 100 leaf switches each with 48 × 25 GbE downlinks (total 120 Tbps) at 3:1 oversubscription yields a bisection bandwidth of 40 Tbps. If workloads demand more east-west bandwidth, the only levers are: reduce the oversubscription ratio (add more spines or upgrade uplinks), add more fabrics (horizontal scaling), or accept that some flows will queue and see elevated latency.

fabrik reports bisection bandwidth as an aggregate across all fabrics in a design, letting you compare different spine count configurations side by side and understand the bandwidth headroom available to your workloads before any congestion occurs.
