package feed

// feed contains the core components of the police feed server.
//
// The server retrieves events from the Swedish Police RSS feed, adding new
// or changed events to the database. Events from the RSS feed are identified
// by their article URL. Since this is not a very performant identifier, a
// UUIDv5 is generated to be used as an internal identifier. This ID should not
// be shared outside the service domain.
//
// The first time an event happens, it is added to the database with revision 1.
// To check if an event has been updated, a content hash is made of the RSS
// event title and description. A new hash means a new event in the table, and
// the revision increases by one.
//
// TODO: all event links in the RSS feed are also crawled so that the article
// contents can also be registered with the event. Due to rate limiting
// concerns, there is a limit to how many events can be crawled at a moment in
// time.
//
// TODO: all events are geocoded based on their contents using provider X.
//
