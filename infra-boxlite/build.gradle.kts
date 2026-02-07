dependencies {
    implementation(project(":core"))
    implementation(fileTree(mapOf("dir" to rootProject.file("libs"), "include" to listOf("*.jar"))))
}
