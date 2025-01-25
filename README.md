# ESGBook Data & Analytics Software Engineer Technical Test

This repository contains the exercise for ESGBook Data & Analytics Software Engineer roles.

## Time

We recommend spending around 1-2 hours on this test. But feel free to spend more time if you want to.

## Requirements

There are two main approaches that people usually take when doing this test:

Either working code, usually in the form of a POC that demonstrates your understanding of the problem and your approach
to solving it.

Or a system design diagram, EDD or ADRs that explain your design decisions, how your solution would work and
what the solution would look like from POC to Production.

We would prefer code as it's easier to understand and evaluate, but we understand that sometimes it's easier to explain
and time constraints might not allow for working code.

For both approaches, we would like to see:

- A brief explanation of the trade-offs you made, and what you would do differently if you had more time.
- A brief explanation of how you would scale your solution if the dataset was 100x larger.
- A brief explanation of how you would deploy, monitor and maintain your solution in production.

## Introduction

Some of our main challenges at ESGBook are related to data processing, data storage, and data analysis. We have a lot of
data coming in from various sources, and we need to process it, store it, and analyze it in a way that is efficient and
scalable.

Off the back of this data processing, we also need to build tools and services that allow product teams to build
products that are data-driven and that can leverage the data we have processed and stored.

These products are mainly scores and insights against companies, usually scoring them on their ESG performance, but also
on other metrics that are important to our clients.

## Objective

For this problem we would like you to build a simple scoring system, something that we internally call config based
scoring. This is a simple system that allows us to score companies based on a set of rules that are defined in a config
file, these configs are usually defined by our product teams and thus are simple to understand and easy to change.

The system should be able to read a config file, process the data, and output a score for each company.

### Data

The data that we have is simple key-value pairs, for this test we have 3 datasets:

- emissions
- waste
- disclosure

You'll be able to find the datasets within the `data` directory, for this test we use CSVs but as you can problably
imagine in production we use a variety of data sources. It would be beneficial if you could think about how you would
handle different types of data sources.

### Score Explanation

A scoring product is a product that is made up of multiple metrics, these metrics can are simple key-value pairs,
as you can see from the `score_1.yaml` file, the product is made up of four metrics `metric_1`, `metric_2`, `metric_3`,
and `metric_4`. Each metric has an operation that defines how the metric is calculated.

In the case of `metric_1`, it's a simple sum of two metrics `waste.was_1` and `disclosure.dis_1`. For this test we
only include 3 types of operations:

- sum; which adds all the parameters together
- or; which accepts 2 parameters (x, y) - if x is null then return y, else return y, if both are null then return null
- divide; which accepts 2 parameters (x, y) which divides x by y.

#### Config Layout

```yaml
name: <name of score>

# list of metrics that are the "product".
metrics:
  - name: metric_1
    operation:
      # there are 3 types of operations for this test: sum, or and divide
      type: sum
      parameters:
        # source is the data source, in this case, it's a simple key-value pair in the format <dataset>.<metric>
        - source: waste.was_1
        # self.<metric> is a special keyword that allows you to reference a metric within the product itself.
        - source: self.metric_2

  - name: metric_2
    operation:
      # or operation is a simple operation that takes two parameters, if x is null then return y, if both are null 
      # then return null
      type: or
      parameters:
        - source: emissions.emi_1
          param: x
        - source: emissions.emi_4
          param: y
```

## Challenges

1) First, you need to write or design a simple scoring system that reads a config file, processes the data, and outputs
   a score for each company for each year. Example output could be:

   ```csv
   company,year,metric_1
   company_1,2020,100
   company_2,2020,200
   company_3,2020,300
   ```

   You'll notice that some of our source data is not in YYYY format, but YYYY-MM-DD format, you should use the latest
   for
   the year for the input data.

   This can be a simple POC or a simple design diagram, we would like to see how you would approach this problem and
   what
   you would do if you had more time.

2) Next, we would like you to think about how you would scale this solution if the dataset was 100x larger. What would
   you do differently, and how would you approach this problem? Imagine that the data is coming in from various sources
   and that the data is being processed in real-time.

3) Finally, we would like you to think about how you would deploy, monitor, and maintain this solution in production.

### Optional Challenges

These are optional challenges that you can do if you have time, or at least think about as we'll discuss them in the
interview.

- How would you handle different types of data sources, from realtime sources ex. CDC, streams to batch sources ex.
  S3, GCS.
- How would you handle different types of data, from structured data to unstructured data.
- How would your system handle data quality issues, ex. missing data, incorrect data, etc.
- How would the system handle scores relying on other scores, ex. score_1.metric_1 had a dependency on score_2.metric_1.
- How would you handle different types of operations, how would you make it extensible and easy to add new operations.

### Submitting

Please send us your solution via email, do not upload it to a public repository.

If you have any questions, please don't hesitate to ask.

### DELETE

Introducing a worker pool in Go can significantly speed up or simplify data processing for large data sets by parallelizing tasks. The idea is:

Break the workload into small “jobs.”
Spawn a fixed number of worker goroutines (the “pool”).
Distribute jobs among the workers via channels.
Collect results back from the workers.
Below is a high-level example showing how you might process (company, year) pairs in parallel when computing scores. You can adapt this pattern to reading CSVs, computing metrics, or any CPU/memory-heavy sub-task.
# Golang-data-filter
