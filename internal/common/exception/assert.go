package exception

// NotNull mirrors org.springframework.util.Assert.notNull(Object, String):
// it throws IllegalArgumentException when the object is null.
func NotNull(obj any, message string) {
	if isNil(obj) {
		panic(&IllegalArgumentException{Message: message})
	}
}

// IsTrue mirrors org.springframework.util.Assert.isTrue(boolean, String).
func IsTrue(expression bool, message string) {
	if !expression {
		panic(&IllegalArgumentException{Message: message})
	}
}

// ThrowNPE reproduces an unguarded dereference in the Java source.
func ThrowNPE() {
	panic(&NullPointerException{})
}
