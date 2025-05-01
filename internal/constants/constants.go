package constants

// DefaultPath for the server
const DefaultPath = "/"

// Version of the API
const Version = "v1"

// Internal Endpoints

const APIPath = DefaultPath + "api/" + Version + "/"
const ForestryRoadsPath = APIPath + "forestryroads"

const ProxyPath = DefaultPath + "proxy/"
const BaseLayerPath = ProxyPath + "baselayer"

// External Endpoints

const NVEFrostDepthAPI = "https://gts.nve.no/api/MultiPointTimeSeries/ByMapCoordinateCsv"
const ForestryRoadsWFS = "https://wms.geonorge.no/skwms1/wms.traktorveg_skogsbilveger"

// SeNorge API themes

const SeNorgeFrostDepthTheme = "gwb_frd"
const SeNorgeWaterSaturationTheme = "gwb_sssrel"
