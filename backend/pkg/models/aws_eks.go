package models

type AWSConfig struct {
	Model
	UserID          uint   `gorm:"not null;index" json:"user_id"`
	Name            string `gorm:"not null" json:"name"`
	AccessKeyID     string `gorm:"not null" json:"access_key_id"`
	SecretAccessKey string `gorm:"not null" json:"-"` // Never expose via JSON
	SessionToken    string `json:"-"`                 // Never expose via JSON
	Region          string `gorm:"not null" json:"region"`
}

type EKSCluster struct {
	Model
	UserID                   uint   `gorm:"not null;index" json:"user_id"`
	AWSConfigID              uint   `gorm:"not null;index" json:"aws_config_id"`
	Name                     string `gorm:"not null" json:"name"`
	Region                   string `gorm:"not null" json:"region"`
	AccountID                string `gorm:"not null" json:"account_id"`
	Endpoint                 string `gorm:"not null" json:"endpoint"`
	CertificateAuthorityData string `gorm:"not null" json:"certificate_authority_data"`

	AWSConfig AWSConfig `gorm:"foreignKey:AWSConfigID" json:"aws_config"`
}
