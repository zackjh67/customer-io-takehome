# Zack's Notes

## Implementation
### Backend
I mostly used what was there for me given that this is the first time I've even looked at Go, and it would have taken even longer to build some of this stuff from scratch for me. The changes I made to the base project structure were obviously implementing the CRUD methods as well as creating a database interface for the application to use internally as well as with endpoints. I chose to use Postgres because I am familiar with it (huge mistake, see below) and the database file basically just houses methods that execute sql queries and return something in the shape that the rest of the app wants. This file also destroys any exisiting database/tables and creates new ones each time you boot up the app. This is not ideal and not the place for that; but it was a simple way of getting this database up and running without having to jump through a bunch of vm setup hoops. You will need to make sure you have Postgres 13 installed and setup to run this app. Let me know if you have any questions at all about that as it has just been a while since I have setup postgres from scratch so there may be a step or 2 that I am missing! 

### Frontend
I am most familiar with Angular and can set an app up the quickest as well as have prior art for it, plus it is my favorite frontend framework so far, so this project was written in Angular/Typescript. There isn't much to the app really because of time constraints, but it basically has simple routes for the different views you may expect (/customers, /customer/:id) which display a list of customers or a single customer given an existing ID in the url. I usually try and get away with not putting redux in smaller projects since it can be a can of worms, and in my opinion it really isn't necessary until you have a medium-large enterprise app, so the frontend relies on Angular services for state and retrieving data. There is an ApiService which does the web calls and a CustomerService that actually interfaces with the ApiService, and in turn components that need customer information call and listen to the CustomerService. The services interact via rxjs streams (Listener/Observer pattern) because that is the Angulary way and for some reason I like them and they are terrible and amazing. 

## Known Issues
### Backend
* Updating customers not implemented. The actual database methods work and are used internally, but this is not yet exposed via the API
* Timestamp support. I cut this for time and planned to get back to it later but didn't have enough time. Difficulties came trying to make Golang and Postgres happy about the same timestamp
* User events. I believe the query is correct to grab the event stats the users need for the payload, but there is something happening where the object queried from the database keeps growing with every user and each subsequent user queried receives their events plus the last user's events. I was not able to fix this in time because it has been years since I have had to work with pointers and never since I have had to work with Go.
* Several CRUD database methods have been stubbed out because of lack of time
* The database creation and table creation belongs somewhere else entirely, and for some reason tables/triggers/functions would persist between sessions (probably forgot to kill something) so it manually drops them and readds them every time.... terrible!
* No pagination support. This is similar to the pagination strategy I use in my current app at Building Engines so it wasn't particularly complicated but I just only had the time to make everything query and return data, so pagination went on the backburner.

### Frontend
* Updating customers not yet implemented. This is actually a relatively easy thing to do using Angular reactive forms but time
* UI is terrible. I usually wait to style things until functionality is done, so it got cut. Flow for me would normally be: figure out how components are going to talk to each other and the backend -> build out views -> get app interfacing with backend -> usually lots of manually testing until everything is working -> componentize what may need it -> make pretty -> write tests for complicated areas of the app
* I cheat a lot in typescript so there are quite a few "any" typed variables
* No accessibility. This is not something I really have experience with but would love to learn!
* No header component with routes for navigation and displaying page title, etc
* No kind of loading indicators whatsoever

## Changes If Only I Had The Time
### Backend
* OMG take a Go crash course before just jumping right in
* Use something a little less rigid than Postgres. The Go Postgres implementation doesn't seem particularly easy to use or intuitive, and they fight about types a lot. It is also just so time consuming for a project like this. A lot of my time was spent just getting Golang and Postgres not to fight the entire time. Plus the API for parsing data out of query results is clunky and insanity to me. I wish I would have done something a little bit easier and more lightweight like Firebase.
* I would have liked to create middleware between the app and the database that basically takes care of all the formatting and such. This would have made it a little more flexible too incase it ended up using multiple databases or maybe down the line I wanted to switch from Postgres
* Data caching
* Build out better types for payloads and such
* Explore different packages that probably make some of the difficult and mundaine things about Golang easy
* Pagination
* Tests

### Frontend
* Actual UI lol
* Explore a state management option somewhere between full-fledged redux and services
* I wanted to use Tailwind and planned to after basic functionality was there but didn't get to :(
* Accessibility... since... you know, the instructions called it out and I didn't get to any of it at all...
* Pagination preferences that persist between sessions
* Loading component
* Banner component
* Main menu component maybe
* Any kind of styling/scss (I promise I know how lol)
* Tests that aren't just the generated Angular ones!
