plugins {
    id("java")
    id("com.github.johnrengelman.shadow") version "8.1.1" // Apply the Shadow plugin
    id("io.freefair.aspectj.post-compile-weaving") version "8.6"
}

group = "org.example"
version = "1.0-SNAPSHOT"

//java { toolchain { languageVersion.set(JavaLanguageVersion.of(17)) } }

repositories {
    mavenCentral()
    maven { url = uri("https://jitpack.io") }
}

dependencies {
    implementation("org.hyperledger.fabric-chaincode-java:fabric-chaincode-shim:2.5.0")
    implementation(files("libs/hypernate-1.0.jar"))
    implementation("org.aspectj:aspectjrt:1.9.7")
    implementation("com.jcabi:jcabi-aspects:0.26.0")
    implementation("org.slf4j:slf4j-simple:2.0.9")
    aspect("com.jcabi:jcabi-aspects:0.26.0")

    testImplementation(platform("org.junit:junit-bom:5.10.0"))
    testImplementation("org.junit.jupiter:junit-jupiter")
    testImplementation("org.slf4j:slf4j-simple:2.0.13")
    testImplementation("org.assertj:assertj-core:3.24.2")
    testImplementation("org.junit.jupiter:junit-jupiter:5.10.0")
    testImplementation("org.mockito:mockito-core:5.11.0")
    testImplementation("org.mockito:mockito-junit-jupiter:5.11.0")
}

tasks.test {
    useJUnitPlatform()
}

// Optional: Configure the shadowJar task if needed
tasks.withType<com.github.jengelman.gradle.plugins.shadow.tasks.ShadowJar> {
    archiveFileName.set("chaincode.jar")
    manifest {
        attributes["Main-Class"] = "org.hyperledger.fabric.contract.ContractRouter"
    }
    from(sourceSets.main.get().output) {
        include("voter/**")
    }
}