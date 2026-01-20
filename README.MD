# Card Barrage

Card Barrage is a MTG Passion project. The stack uses Express JS, a Go back end, with a dual databases to power a suggestion engine. The front end is a simple Express JS webapp, which uses NextJS for auth. The back end utilizes Go to allow for some powerful low level feautres and multi threading, though very little has been utilized. The dual databases power a suggestion engine. The single source of truth is a Postgre database accessed via GORM. The suggestion engine lives in a Memgraph database that stores only fundamental gameplay data.
