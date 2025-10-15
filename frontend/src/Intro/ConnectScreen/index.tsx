import { Button, Card, Divider } from '@blueprintjs/core';
import { Trans, useTranslation } from 'react-i18next';

import { BrowserOpenURL } from '../../../wailsjs/runtime/runtime';
import BlueSkyLogo from '../../assets/icons/bluesky.svg';
import DiscordIcon from '../../assets/icons/discord.svg';
import GithubIcon from '../../assets/icons/github.svg';
import OpenCollectiveIcon from '../../assets/icons/osc.svg';
import RedditIcon from '../../assets/icons/reddit.svg';
import { SOCIAL_LINKS } from '../../constants/urls';

import './index.css';

export function ConnectScreen() {
  const { t } = useTranslation();

  return (
    <div className="intro-screen">
      <h3 className="bp5-heading intro-heading">{t('intro.connect.title')}</h3>
      <p className="intro-description bp5-running-text">{t('intro.connect.description')}</p>
      <p className="intro-description bp5-running-text">
        <Trans
          i18nKey="intro.connect.caNote"
          components={{
            strong: <strong />,
            br: <br />,
          }}
        />
      </p>

      <Card elevation={1} className="connect-card">
        <p className="bp5-heading">
          <strong>{t('intro.connect.socialText')}</strong>
        </p>

        <div className="social-links-grid">
          <div className="social-row">
            <Button fill onClick={() => BrowserOpenURL(SOCIAL_LINKS.GITHUB)} className="social-button">
              <img src={GithubIcon} className="social-icon" alt="GitHub" />
              GitHub
            </Button>

            <Button fill onClick={() => BrowserOpenURL(SOCIAL_LINKS.BLUESKY)} className="social-button">
              <img src={BlueSkyLogo} className="social-icon" alt="Bluesky" />
              Bluesky
            </Button>
          </div>

          <div className="social-row">
            <Button fill onClick={() => BrowserOpenURL(SOCIAL_LINKS.REDDIT)} className="social-button">
              <img src={RedditIcon} className="social-icon" alt="Reddit" />
              Reddit
            </Button>

            <Button fill onClick={() => BrowserOpenURL(SOCIAL_LINKS.DISCORD)} className="social-button">
              <img src={DiscordIcon} className="social-icon" alt="Discord" />
              Discord
            </Button>
          </div>
        </div>

        <Divider className="section-divider" />

        <p>{t('intro.connect.donateText')}</p>
        <Button
          icon={<img src={OpenCollectiveIcon} className="social-icon" alt="Open Collective" />}
          onClick={() => BrowserOpenURL(SOCIAL_LINKS.OPEN_COLLECTIVE)}
        >
          Open Collective
        </Button>
      </Card>
    </div>
  );
}
