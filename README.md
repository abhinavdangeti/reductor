# reductor
This library/tool aims at reducing the footprint of a postings list (an array of uint32s).

## how does it work
Reductor assumes that the provided postings list is already sorted. It estimates the difference or the delta between adjacent entries and stores the deltas which are much smaller values than the original entries. These deltas are then bitpacked to form a highly compresses data structure carrying all the necessary information needed to rebuild the original postings list.

## services offered
As of now, reductor offers 2 functional APIs - one to encode a provided list (AddAll), two to decode a provided list (FetchAll).
