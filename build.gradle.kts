import org.jetbrains.kotlin.gradle.dsl.JvmTarget
import org.jetbrains.kotlin.gradle.tasks.KotlinCompile

plugins {
    kotlin("jvm") version "2.3.0" apply false
}

group = "com.coooolfan"
version = "1.0.0"

allprojects {
    repositories {
        mavenLocal()
        mavenCentral()
        maven { url = uri("https://mirrors.cloud.tencent.com/nexus/repository/maven-public/") }
    }
}

subprojects {
    apply(plugin = "org.jetbrains.kotlin.jvm")

    dependencies {
        "testImplementation"(kotlin("test"))
        "testRuntimeOnly"("org.junit.platform:junit-platform-launcher")
    }

    tasks.withType<Test>().configureEach {
        useJUnitPlatform()
    }

    tasks.withType<JavaCompile>().configureEach {
        options.encoding = "UTF-8"
        options.compilerArgs.add("-parameters")
        options.release.set(25)
    }

    tasks.withType<KotlinCompile>().configureEach {
        compilerOptions {
            javaParameters.set(true)
            jvmTarget.set(JvmTarget.JVM_25)
            freeCompilerArgs.add("-Xjsr305=strict")
        }
    }
}
