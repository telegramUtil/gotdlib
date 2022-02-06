## Example

### Bot
Login to bot account.

### Command
Handle user command and reply it.

### Event Filter
Since we can have many update type in updates.
So we need to filter update events, like UpdateNewMessage, UpdateMessageSendSucceeded, UpdateMessageSendFailed, etc.

### Media
Send photo or album to chat.

### Pending updates
When starting a bot, we may have some updates that are missed to process when a listener IS NOT ready.

So we need to keep specific update types in memory until a listener is set, then we can process those updates again.

### Raw Update
Get update without event filter.