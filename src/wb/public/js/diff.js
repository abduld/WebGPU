$(function() {
	var diffView = $("#diff-view");
	if (diffView.length > 0) {
		diffView = diffView[0];
		window.editor = CodeMirror.MergeView(diffView, {
	      lineNumbers: true,
	      tabSize: 4,
	      indentUnit: 4,
	      indentWithTabs: true,
	      readOnly: true,
	      value: $("#program-value").text(),
	      orig: $("#program-orig").text(),
	      highlightDifferences: true,
		  mode: "text/x-c++src",
		});
	}
});
