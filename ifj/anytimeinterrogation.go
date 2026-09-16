package ifj

/*
anyTimeInterrogation  OPERATION ::= {	--Timer m
	ARGUMENT
		AnyTimeInterrogationArg
	RESULT
		AnyTimeInterrogationRes
	ERRORS {
		systemFailure       |
		ati-NotAllowed      |
		dataMissing         |
		unexpectedDataValue |
		unknownSubscriber}
	CODE	local:71 }
*/

/*
AnyTimeInterrogationArg ::= SEQUENCE {
	subscriberIdentity	[0] SubscriberIdentity,
	requestedInfo	    [1] RequestedInfo,
	gsmSCF-Address      [3] ISDN-AddressString,
	extensionContainer  [2] ExtensionContainer  OPTIONAL,
	...}
*/

/*
SubscriberIdentity ::= CHOICE {
	imsi	[0] IMSI,
	msisdn	[1] ISDN-AddressString
	}
*/

/*
RequestedInfo ::= ENUMERATED {
	anchorMSC-AddressAndASCI-CallReference         (0),
	imsiAndAdditionalInfoAndAdditionalSubscription (1),
	... }
--	exception handling:
--	an unrecognized value shall be rejected by the receiver with a return error cause of
--	unexpected data value
*/

/*
AnyTimeInterrogationRes ::= SEQUENCE {
	subscriberInfo	    SubscriberInfo,
	extensionContainer	ExtensionContainer	OPTIONAL,
	...}
*/

/*
SubscriberInfo ::= SEQUENCE {
	locationInformation	[0] LocationInformation OPTIONAL,
	subscriberState     [1] SubscriberState     OPTIONAL,
	extensionContainer  [2] ExtensionContainer  OPTIONAL,
	^^^ R99 ^^^
	... ,
	locationInformationGPRS [3] LocationInformationGPRS OPTIONAL,
	ps-SubscriberState      [4] PS-SubscriberState      OPTIONAL,
	imei                    [5] IMEI                    OPTIONAL,
	ms-Classmark2           [6] MS-Classmark2           OPTIONAL,
	gprs-MS-Class           [7] GPRSMSClass             OPTIONAL,
	mnpInfoRes              [8] MNPInfoRes              OPTIONAL,
	imsVoiceOverPS-SessionsIndication  [9] IMS-VoiceOverPS-SessionsInd OPTIONAL,
	lastUE-ActivityTime     [10] Time                   OPTIONAL,
	lastRAT-Type            [11] Used-RAT-Type          OPTIONAL,
	eps-SubscriberState     [12] PS-SubscriberState     OPTIONAL,
	locationInformationEPS  [13] LocationInformationEPS OPTIONAL,
	timeZone                [14] TimeZone               OPTIONAL,
	daylightSavingTime      [15] DaylightSavingTime     OPTIONAL,
	locationInformation5GS  [16] LocationInformation5GS OPTIONAL }

--	If the HLR receives locationInformation, subscriberState or ms-Classmark2 from an SGSN or
--	MME (via an IWF), it shall discard them.
--	If the HLR receives locationInformationGPRS, ps-SubscriberState, gprs-MS-Class or
--	locationInformationEPS (outside the locationInformation IE) from a VLR, it shall
--	discard them.
--	If the HLR receives parameters which it has not requested, it shall discard them.
--	The locationInformation5GS IE should be absent if UE did not access via 5GS and IM-SSF.
*/

/*
LocationInformation ::= SEQUENCE {
	ageOfLocationInformation	AgeOfLocationInformation	OPTIONAL,
	geographicalInformation	[0] GeographicalInformation	OPTIONAL,
	vlr-number	[1] ISDN-AddressString	OPTIONAL,
	locationNumber	[2] LocationNumber	OPTIONAL,
	cellGlobalIdOrServiceAreaIdOrLAI	[3] CellGlobalIdOrServiceAreaIdOrLAI	OPTIONAL,
	extensionContainer	[4] ExtensionContainer	OPTIONAL,
	... ,
	selectedLSA-Id	[5] LSAIdentity	OPTIONAL,
	msc-Number	[6] ISDN-AddressString	OPTIONAL,
	geodeticInformation	[7] GeodeticInformation	OPTIONAL,
	currentLocationRetrieved	[8] NULL	OPTIONAL,
	sai-Present	[9] NULL	OPTIONAL,
	locationInformationEPS	[10] LocationInformationEPS	OPTIONAL,
	userCSGInformation	[11] UserCSGInformation	OPTIONAL }
-- sai-Present indicates that the cellGlobalIdOrServiceAreaIdOrLAI parameter contains
-- a Service Area Identity.
-- currentLocationRetrieved shall be present
-- if the location information were retrieved after a successfull paging.
-- if the locationinformationEPS IE is present then the cellGlobalIdOrServiceAreaIdOrLAI IE,
-- the ageOfLocationInformation IE, the geographicalInformation IE, the geodeticInformation IE
-- and the currentLocationRetrieved IE (outside the locationInformationEPS IE) shall be
-- absent. As an exception, both the cellGlobalIdOrServiceAreaIdOrLAI IE including an LAI and
-- the locationinformationEPS IE may be present in a MAP-NOTE-MM-EVENT.
-- UserCSGInformation contains the CSG ID, Access mode, and the CSG Membership Indication in
-- the case the Access mode is Hybrid Mode.
-- The locationInformationEPS IE should be absent if locationInformationEPS-Supported was not
-- received in the RequestedInfo IE.
*/

/*
ati-NotAllowed  ERROR ::= {
	PARAMETER
	ATI-NotAllowedParam
	-- optional
	CODE	local:49 }
*/
