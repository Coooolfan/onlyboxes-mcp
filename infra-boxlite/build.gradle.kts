import org.gradle.api.file.DuplicatesStrategy
import org.gradle.jvm.tasks.Jar

val boxliteJarCandidates = rootProject.file("libs")
    .listFiles()
    ?.filter { jarFile ->
        jarFile.isFile &&
            jarFile.name.startsWith("boxlite-java-highlevel-allplatforms-") &&
            jarFile.name.endsWith(".jar") &&
            !jarFile.name.endsWith("-sources.jar")
    }
    .orEmpty()

val boxliteJar = boxliteJarCandidates.maxByOrNull { it.lastModified() }
    ?: throw GradleException(
        "No boxlite highlevel jar found under ${rootProject.file("libs")}." +
            " Expected boxlite-java-highlevel-allplatforms-<version>.jar",
    )

val boxliteVersion = Regex("boxlite-java-highlevel-allplatforms-(.+)\\.jar")
    .matchEntire(boxliteJar.name)
    ?.groupValues
    ?.get(1)
    ?: throw GradleException("Cannot parse boxlite version from jar name: ${boxliteJar.name}")

val sanitizedBoxliteJar by tasks.registering(Jar::class) {
    group = "build"
    description = "Repackages boxlite jar and strips embedded Jackson classes to avoid runtime conflicts"

    archiveBaseName.set("boxlite-java-highlevel-allplatforms")
    archiveVersion.set(boxliteVersion)
    archiveClassifier.set("sanitized")
    destinationDirectory.set(layout.buildDirectory.dir("sanitized-libs"))

    duplicatesStrategy = DuplicatesStrategy.EXCLUDE

    from(zipTree(boxliteJar))

    // Keep one Jackson stack on the app classpath. Spring Boot manages those versions.
    exclude("com/fasterxml/jackson/**")
    exclude("META-INF/maven/com.fasterxml.jackson.core/**")
    exclude("META-INF/versions/*/com/fasterxml/jackson/**")
    exclude("META-INF/*.SF")
    exclude("META-INF/*.DSA")
    exclude("META-INF/*.RSA")
}

dependencies {
    implementation(project(":core"))
    implementation(files(sanitizedBoxliteJar.flatMap { it.archiveFile }))
}
