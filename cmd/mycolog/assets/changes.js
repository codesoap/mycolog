let unsubmittedChanges = false;

window.addEventListener('beforeunload', (event) => {
	if (unsubmittedChanges) {
		event.preventDefault();
		event.returnValue = ''; // Chrome requires returnValue to be set.
	}
});

function registerChange() {
	unsubmittedChanges = true;
}

function registerSubmit() {
	unsubmittedChanges = false;
}
