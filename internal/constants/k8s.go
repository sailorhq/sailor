package constants

const (
	CONFIG_VOLUME_NAME = "sailor-config-volume"
	SECRET_VOLUME_NAME = "sailor-secret-volume"
	MISC_VOLUME_NAME   = "sailor-misc-volume"

	CONFIG_VOLUME_PATH = "/etc/sailor"
	SECRET_VOLUME_PATH = "/etc/sailor/secret"
	MISC_VOLUME_PATH   = "/etc/sailor/misc"

	// Labels
	LABEL_ADMISSION   = "sailorhq.dev/admission"
	LABEL_RELEASE_TAG = "sailorhq.dev/release-tag"
)
