# Customer.io Challenge README

## Part One: Summarize and serve data

For this project, you will write an application that summarizes and provides an interface to view two types of **user behavioural data** stored in a JSON-encoded file. This file contains one activity item per line, each tied to a **user_id**, of the following types:

- **One-time events**, which represent activities performed by a user at specific point in time
- **Attribute changes,** which represent the setting of **persistent attribute values** for the user at a specific point in time

**Example Events:**

```
{"id":"c7d1a8d9-da03-11e4-87ec-946849a0cf6a","type":"event","name":"page","user_id":"2352","data":{"url": "http://mystore.com/product/socks"},"timestamp":1428067050}
{"id":"735a247d-7179-5024-1686-ab353a730b45","type":"event","name":"purchase","user_id":"2352","data":{"sku": "CMR01", "price": "19.99"},"timestamp":1428067050}
```

**Example attribute change:**

```
{"id":"c52543d8-da03-11e4-8e29-c5dc2fe5941b","type":"attributes","user_id":"2352","data":{"name": "Bill", "email": "bill@gmail.com"},"timestamp":1428067050}
```

Using the code in this **.tar** file as a base, your program should iterate through each line of an input file and for each unique **user_id** present in the file:

- Keep a record of the **latest value** for each set attribute, where latest means the **most recent timestamp specified in an attribute message for a user_id/attribute pair.** The attributes being set are stored in the **data** hash of the attribute change message
- Keep a count of the **unique number of times a given event type was performed** for this user_id, taking into account the possibility of duplicated IDs

This summarization of the data is what will power the user interface. To make this work, you can implement the `serve.Datastore` interface, we've included a skeleton in `datastore/datastore.go` to get you started. This is used by our `serve` package to provide a REST-ful interface you can use for the second portion of the challenge.

**Note** You do not need to use the `serve.Datastore` interface to handle populating your summaries, it is intended to be a layer ontop of your datastore to expose data in a format our server understands, and is not necessary for the summarization step. We generally recommend implementing the summary part first and then writing your `serve.Datastore` implementation afterwards, as the datastore is optimized for server usage patterns and not for summarization.

## Part Two: Add a User Interface

The goal for part two of this homework is to build a UI to list the customer records summarized in the previous exercise, and be able to display and edit a single customer's attributes. The UI you build should communicate with the REST-ful API provided by the `serve` package and included in `main.go` (see specification below).

The UI we build at Customer.io is written in Ember.js, but for this homework feel free to choose any framework or tools you're most comfortable with.


### UI Technical Background

The interface you'll build is inspired by parts of our real app, where our users can view or edit the records of their customers. Here are some assumptions we make about the customer records:

1. Each customer has a unique and immutable id (`id`), an email address (`email`) and a `created_at` attribute
2. A customer may also have any number of other attributes (coming from the attribute changes from Part One, or by making manual changes via the UI you're building)
3. Each customer may have a different set of attributes from other customers
4. Because of how we store customer data, attributes are returned from the API as a nested JSON object inside a customer record
5. To prevent last-write-wins scenarios, attribute changes are merged, which makes deleting attributes a little unusual

In practice, our users will sync their customer database with ours, either using our REST API (through attribute changes), our JavaScript snippet, or a third-party integration like Segment.com. But we still want to be able to browse and manage customer records in our app.

#### UI views to build
There are three pages to build. Below you'll find design mockups as a guideline - your final app doesn't have to look exactly like them. Feel free to use Bootstrap, Tailwind or any other CSS frameworks in your implementation.

**List all customers, with a link to each customer’s view and showing basic info**
![image3](https://user-images.githubusercontent.com/3127419/116959729-dec4f180-ace1-11eb-8dfb-77efbb64c589.png)

**Show customer, showing basic info, a list of attributes, a list of events, and a link to edit attributes**
![customer](https://user-images.githubusercontent.com/3127419/122132168-b4885700-ce7d-11eb-8bbf-3feb522e5c17.png)

**Note:** The event list is optional - while a summary of a customer's events are included in the API payload, you don't have to use them. Editing attributes should be a higher priority in this task. 

**Edit customer, which will support adding, removing, and changing attributes**
![image2](https://user-images.githubusercontent.com/3127419/116959739-e71d2c80-ace1-11eb-8d1d-55a6a0fa801e.png)

The exercise shouldn't take more than a few hours of your time. If you don't complete the entire exercise, that's okay! We don't expect anyone to deliver a polished, fully-working solution. The next step after this exercise is a code review & pairing interview, where we talk through what problems you faced, and what your next steps would be.

Be creative, if you're so inclined! At Customer.io everyone has an impact on the product, so your input on what we should be doing in a feature team is valuable. The mockups are guidelines, and adhering to them is not required. Show us your ideas.

Finally, remember that communication and documentation is just as important as code. Please write a few bullet points about your implementation, any known bugs, and things you'd change given more time. And if you have any questions, just drop us an email!


### REST API Specification

We've provided you with an implementation of the REST server that works with the `Datastore` in the `main.go` file, it listens on `localhost:1323`, and exposes the following endpoints

<hr>


`GET localhost:1323/customers` - retrieve a list of customers, paginated. Accepts two query params `?page=N&per_page=M`. Page starts at 1

### example response:
```
{
  "customers": [
    {
      "id": 1004,
      "attributes": {
        "created_at": "1542474417",
        "email": "chloemoore441@example.com",
        "first_name": "Sofia",
        "ip": "114.32.23.98",
        "last_name": "Thomas",
      },
      "events": {
        "purchase": 2,
        "page": 6
      },
      "last_updated": 1560964022
    },
    {
      "id": 10040,
      "attributes": {
        "city": "Hoonah",
        "created_at": "1550682417",,
        "email": "ethananderson130@test.org",
      },
      "events": {
        "page": 1
      },
      "last_updated": 1560964021
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 25,
    "total": 2
  }
}
```
<hr>

`GET localhost:1323/customers/:id` - retrieve a single customer

### example response

```
{
  "customer": {
    "id": 1004,
    "attributes": {
        "created_at": "1542474417",
        "email": "chloemoore441@example.com",
        "first_name": "Sofia",
        "ip": "114.32.23.98",
        "last_name": "Thomas",
    },
    "events": {
      "purchase": 1,
      "page": 3
    },
    "last_updated": 1560964022
  }
}
```
<hr>

`DELETE localhost:1323/customers/:id` - delete a customer by ID. Returns a `201` response on success

<hr>

`POST localhost:1323/customers` - create a customer

### example request body

```
{
    "customer": {
        "id": 12345,
        "attributes": {
            "created_at": "1560964022",
            "email": "example@customer.io",
            "first_name": "example"
        }
    }
}
```

### example response body

```
{
    "customer": {
        "id": 12345,
        "attributes": {
            "created_at": "1560964022",
            "email": "example@customer.io",
            "first_name": "example"
        },
        "events": {},
        "last_updated": 1620135856
    }
}
```
<hr>

`PATCH localhost:1323/customers/:id` - update a customer

#### example request body

```
{
    "customer": {
        "attributes": {
            "ip": "127.0.0.1",
            "first_name": "real",
            "last_name": "customer"
        }
    }
}
```


#### example response body

```
{
    "customer": {
        "id": 12345,
        "attributes": {
            "created_at": "1560964022",
            "email": "example@customer.io",
            "first_name": "real",
            "ip": "127.0.0.1",
            "last_name": "customer"
        },
        "events": {
          "purchase": 1,
          "page": 3
        },
        "last_updated": 1620135856
    }
}
```

## Setting up your environment

You should be using **go 1.15 or later** using go modules (go.mod file provided)

We recommend developing using Visual Studio Code https://code.visualstudio.com/ and the vscode-go plugin https://github.com/Microsoft/vscode-go or a similar configuration. This will set up some of the standard tooling you need to get started with a go project. Feel free to use external packages, which you can install with `go get packagename`.

Some useful resources for getting started:
- Effective Go: https://golang.org/doc/effective_go.html 
- Go Styleguide: https://github.com/golang/go/wiki/CodeReviewComments

But don't worry about having perfectly styled Go, especially if you're new to the language. We're more interested in the functionality of your solution than the specifics of style.

This **.tar** file contains the following:

- A skeleton main program `main.go` which reads and parses the input file line by line, providing you with a channel that you can pull records from, and a server to serve the API. You can choose to use this, but it's fine if you'd prefer to write your own. To run the main program you can use `go run main.go`. Once it is finished summarizing, by default it will serve the results at `localhost:1323`
- A program which you can use to generate test data. The `generate/main.go` program generates two files: a .**data** file, which contains the input JSON data and a **.csv** file which contains a sample summarization of the data.
- A program you can use to verify your summary data. The `verify/main.go` program loads data from the generated summarization file above and compares it to results from your API.

 We recommend generating two test datasets in the **data/** directory as follows:

```
go run generate/main.go -out data/messages.1.data -verify data/verify.1.csv --seed 1560981440 -count 20
go run generate/main.go -out data/messages.2.data -verify data/verify.2.csv --seed 1560980000 -count 10000 -attrs 20 -events 300000 -maxevents 500 -dupes 10
```

After you've finished implementing part one of the solution, you can verify it against the generated data using the verify utility. Note that you'll need to have your summary server running for this script to verify your results, as it will perform requests to the customer list endpoint to retrieve your summarized customer data.

```
go run verify/main.go --verify-file=/path/to/verify/file
```

## Evaluation Criteria

### For part one, here's what we're looking for:

1. A solution that simply and reliably summarizes input data. Correctness and readability are key requirements here.
2. A full implementation of the `Datastore` interface, exposed using the supplied REST API in the `serve` packaged, for interacting with your summarized data. 
3. Your solution can run entirely in memory, but will be run on commodity hardware and should be able to perform reasonably well for a large data file

### For part two, there are three things we're looking for you to deliver:

1. Code: implementation of the list, show & edit customer pages, based on the mockups
2. Tests: test crucial parts of the functionality, full test coverage not required
3. Communication: a few bullet points on known issues, future work, anything else


## Things we'll want to talk about

- Why did you choose the architecture used in your solution?
- What other architectures could you envision for this problem?
- What assumptions did you make?
- Where are the bottlenecks? What's using the most memory, the most cpu, the most time?
- How would you improve the performance?
- How would you improve the UX of the frontend?
- How do you make the UI accessible?
- After these are covered we'll want to discuss how to extend your solution in various ways.
