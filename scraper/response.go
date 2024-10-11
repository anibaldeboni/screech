package scraper

type GameInfoResponse struct {
	Header struct {
		APIversion       string `json:"APIversion"`
		DateTime         string `json:"dateTime"`
		CommandRequested string `json:"commandRequested"`
		Success          string `json:"success"`
		Error            string `json:"error"`
	} `json:"header"`
	Response struct {
		Serveurs struct {
			CPU1                  string `json:"cpu1"`
			CPU2                  string `json:"cpu2"`
			CPU3                  string `json:"cpu3"`
			CPU4                  string `json:"cpu4"`
			Threadsmin            string `json:"threadsmin"`
			Nbscrapeurs           string `json:"nbscrapeurs"`
			Apiacces              string `json:"apiacces"`
			Closefornomember      string `json:"closefornomember"`
			Closeforleecher       string `json:"closeforleecher"`
			Maxthreadfornonmember string `json:"maxthreadfornonmember"`
			Threadfornonmember    string `json:"threadfornonmember"`
			Maxthreadformember    string `json:"maxthreadformember"`
			Threadformember       string `json:"threadformember"`
		} `json:"serveurs"`
		Ssuser struct {
			ID                  string `json:"id"`
			Numid               string `json:"numid"`
			Niveau              string `json:"niveau"`
			Contribution        string `json:"contribution"`
			Uploadsysteme       string `json:"uploadsysteme"`
			Uploadinfos         string `json:"uploadinfos"`
			Romasso             string `json:"romasso"`
			Uploadmedia         string `json:"uploadmedia"`
			Propositionok       string `json:"propositionok"`
			Propositionko       string `json:"propositionko"`
			Quotarefu           string `json:"quotarefu"`
			Maxthreads          string `json:"maxthreads"`
			Maxdownloadspeed    string `json:"maxdownloadspeed"`
			Requeststoday       string `json:"requeststoday"`
			Requestskotoday     string `json:"requestskotoday"`
			Maxrequestspermin   string `json:"maxrequestspermin"`
			Maxrequestsperday   string `json:"maxrequestsperday"`
			Maxrequestskoperday string `json:"maxrequestskoperday"`
			Visites             string `json:"visites"`
			Datedernierevisite  string `json:"datedernierevisite"`
			Favregion           string `json:"favregion"`
		} `json:"ssuser"`
		Jeu struct {
			Rom struct {
				ID              string `json:"id"`
				Romnumsupport   string `json:"romnumsupport"`
				Romtotalsupport string `json:"romtotalsupport"`
				Romfilename     string `json:"romfilename"`
				Romtype         string `json:"romtype"`
				Romsupporttype  string `json:"romsupporttype"`
				Romsize         string `json:"romsize"`
				Romcrc          string `json:"romcrc"`
				Rommd5          string `json:"rommd5"`
				Romsha1         string `json:"romsha1"`
				Romcloneof      string `json:"romcloneof"`
				Beta            string `json:"beta"`
				Demo            string `json:"demo"`
				Proto           string `json:"proto"`
				Trad            string `json:"trad"`
				Hack            string `json:"hack"`
				Unl             string `json:"unl"`
				Alt             string `json:"alt"`
				Best            string `json:"best"`
				Netplay         string `json:"netplay"`
			} `json:"rom"`
			Systeme struct {
				ID   string `json:"id"`
				Text string `json:"text"`
			} `json:"systeme"`
			Editeur struct {
				ID   string `json:"id"`
				Text string `json:"text"`
			} `json:"editeur"`
			Developpeur struct {
				ID   string `json:"id"`
				Text string `json:"text"`
			} `json:"developpeur"`
			ID      string `json:"id"`
			Romid   string `json:"romid"`
			Notgame string `json:"notgame"`
			Cloneof string `json:"cloneof"`
			Joueurs struct {
				Text string `json:"text"`
			} `json:"joueurs"`
			Note struct {
				Text string `json:"text"`
			} `json:"note"`
			Topstaff string `json:"topstaff"`
			Rotation string `json:"rotation"`
			Noms     []struct {
				Region string `json:"region"`
				Text   string `json:"text"`
			} `json:"noms"`
			Synopsis []struct {
				Langue string `json:"langue"`
				Text   string `json:"text"`
			} `json:"synopsis"`
			Dates []struct {
				Region string `json:"region"`
				Text   string `json:"text"`
			} `json:"dates"`
			Genres []struct {
				ID         string `json:"id"`
				Nomcourt   string `json:"nomcourt"`
				Principale string `json:"principale"`
				Parentid   string `json:"parentid"`
				Noms       []struct {
					Langue string `json:"langue"`
					Text   string `json:"text"`
				} `json:"noms"`
			} `json:"genres"`
			Familles []struct {
				ID         string `json:"id"`
				Nomcourt   string `json:"nomcourt"`
				Principale string `json:"principale"`
				Parentid   string `json:"parentid"`
				Noms       []struct {
					Langue string `json:"langue"`
					Text   string `json:"text"`
				} `json:"noms"`
			} `json:"familles"`
			Numeros []struct {
				ID         string `json:"id"`
				Nomcourt   string `json:"nomcourt"`
				Principale string `json:"principale"`
				Parentid   string `json:"parentid"`
				Noms       []struct {
					Langue string `json:"langue"`
					Text   string `json:"text"`
				} `json:"noms"`
			} `json:"numeros"`
			Themes []struct {
				ID         string `json:"id"`
				Nomcourt   string `json:"nomcourt"`
				Principale string `json:"principale"`
				Parentid   string `json:"parentid"`
				Noms       []struct {
					Langue string `json:"langue"`
					Text   string `json:"text"`
				} `json:"noms"`
			} `json:"themes"`
			Medias []Media `json:"medias"`
			Roms   []struct {
				ID              string `json:"id"`
				Romsize         string `json:"romsize"`
				Romfilename     string `json:"romfilename"`
				Romnumsupport   string `json:"romnumsupport"`
				Romtotalsupport string `json:"romtotalsupport"`
				Romcloneof      string `json:"romcloneof"`
				Romcrc          string `json:"romcrc"`
				Rommd5          string `json:"rommd5"`
				Romsha1         string `json:"romsha1"`
				Beta            string `json:"beta"`
				Demo            string `json:"demo"`
				Proto           string `json:"proto"`
				Trad            string `json:"trad"`
				Hack            string `json:"hack"`
				Unl             string `json:"unl"`
				Alt             string `json:"alt"`
				Best            string `json:"best"`
				Netplay         string `json:"netplay"`
				Regions         struct {
					RegionsID        []string `json:"regions_id"`
					RegionsShortname []string `json:"regions_shortname"`
					RegionsEn        []string `json:"regions_en"`
					RegionsFr        []string `json:"regions_fr"`
					RegionsDe        []string `json:"regions_de"`
					RegionsEs        []string `json:"regions_es"`
					RegionsPt        []string `json:"regions_pt"`
				} `json:"regions,omitempty"`
				Langues struct {
					LanguesID        []string `json:"langues_id"`
					LanguesShortname []string `json:"langues_shortname"`
					LanguesEn        []string `json:"langues_en"`
					LanguesFr        []string `json:"langues_fr"`
					LanguesDe        []string `json:"langues_de"`
					LanguesEs        []string `json:"langues_es"`
					LanguesIt        []string `json:"langues_it"`
					LanguesPt        []string `json:"langues_pt"`
				} `json:"langues,omitempty"`
			} `json:"roms"`
		} `json:"jeu"`
	} `json:"response"`
}

type Media struct {
	Type      string `json:"type"`
	Parent    string `json:"parent"`
	URL       string `json:"url"`
	Region    string `json:"region,omitempty"`
	Crc       string `json:"crc"`
	Md5       string `json:"md5"`
	Sha1      string `json:"sha1"`
	Size      string `json:"size,omitempty"`
	Format    string `json:"format"`
	Posx      string `json:"posx,omitempty"`
	Posy      string `json:"posy,omitempty"`
	Posw      string `json:"posw,omitempty"`
	Posh      string `json:"posh,omitempty"`
	ID        string `json:"id,omitempty"`
	Subparent string `json:"subparent,omitempty"`
}
