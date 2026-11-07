import java.util.Properties
import java.util.Base64

plugins {
    id("com.android.application")
    id("dev.flutter.flutter-gradle-plugin")
}

val dartDefines = providers.gradleProperty("dart-defines").orNull.orEmpty()
    .split(",").filter { it.isNotEmpty() }
    .associate {
        val decoded = String(Base64.getDecoder().decode(it), Charsets.UTF_8)
        decoded.substringBefore("=") to decoded.substringAfter("=", "")
    }
val allSources = dartDefines["ALL_SOURCES"] == "true"
val tvBuild = dartDefines["TV_BUILD"] == "true"

val releaseKey = rootProject.file("key.properties")
val releaseProperties = Properties()
if (releaseKey.exists()) {
    releaseKey.inputStream().use { releaseProperties.load(it) }
} else {
    mapOf(
        "storeFile" to "ANDROID_KEYSTORE_PATH",
        "storePassword" to "ANDROID_KEYSTORE_PASSWORD",
        "keyAlias" to "ANDROID_KEY_ALIAS",
        "keyPassword" to "ANDROID_KEY_PASSWORD",
    ).forEach { (key, name) -> System.getenv(name)?.let { releaseProperties[key] = it } }
}
val hasReleaseKey = releaseProperties.getProperty("storeFile") != null

android {
    namespace = "com.guoapp"
    compileSdk = flutter.compileSdkVersion
    ndkVersion = "28.2.13676358"

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    defaultConfig {
        applicationId = if (tvBuild) "com.duanju.duanju_app.coexist.tv" else if (allSources) "com.duanju.duanju_app.coexist" else "com.guoapp.phoenix"
        minSdk = 24
        targetSdk = flutter.targetSdkVersion
        // Flutter --split-per-abi adds 2000 to the ARM64 version code.
        // Base 113 produces ARM64 coexist versionCode 2113, upgrading from 2112.
        versionCode = if (tvBuild) 114 else if (allSources) 113 else flutter.versionCode
        versionName = if (tvBuild) "1.0.4+104coexist-TV" else if (allSources) "1.0.4+104coexist" else flutter.versionName
        manifestPlaceholders["appLabel"] = if (tvBuild) "真果鑑 TV" else if (allSources) "真果鉴-共存" else "红果鉴"
        manifestPlaceholders["appBanner"] = if (allSources) "@drawable/tv_banner_all_sources" else "@drawable/tv_banner"
    }

    signingConfigs {
        if (hasReleaseKey) {
            create("release") {
                val rawStore = requireNotNull(releaseProperties.getProperty("storeFile"))
                storeFile = if (file(rawStore).exists()) file(rawStore) else rootProject.file(rawStore)
                storePassword = requireNotNull(releaseProperties.getProperty("storePassword"))
                keyAlias = requireNotNull(releaseProperties.getProperty("keyAlias"))
                keyPassword = requireNotNull(releaseProperties.getProperty("keyPassword"))
            }
        }
    }

    buildTypes {
        debug {
            signingConfig = signingConfigs.getByName(if (hasReleaseKey) "release" else "debug")
        }
        release {
            signingConfig = signingConfigs.getByName(if (hasReleaseKey) "release" else "debug")
        }
        getByName("profile") {
            signingConfig = signingConfigs.getByName(if (hasReleaseKey) "release" else "debug")
        }
    }

    packaging {
        jniLibs {
            useLegacyPackaging = true
        }
    }
}

kotlin {
    compilerOptions {
        jvmTarget = org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17
    }
}

flutter { source = "../.." }

tasks.withType<JavaCompile>().configureEach {
    if (name.contains("Release")) {
        doFirst {
            val registrant = file("src/main/java/io/flutter/plugins/GeneratedPluginRegistrant.java")
            if (registrant.exists()) {
                val generated = registrant.readText()
                val integrationPlugin = Regex(
                    """(?s)    try \{\s*flutterEngine\.getPlugins\(\)\.add\(new dev\.flutter\.plugins\.integration_test\.IntegrationTestPlugin\(\)\);\s*\} catch \(Exception e\) \{[^}]*\}\s*"""
                )
                val release = generated.replace(integrationPlugin, "")
                if (release != generated) registrant.writeText(release)
            }
        }
    }
}
