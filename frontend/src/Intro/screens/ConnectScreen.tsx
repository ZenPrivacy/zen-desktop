import { Button, Card, Divider, H5 } from '@blueprintjs/core';
import { useTranslation } from 'react-i18next';

import { BrowserOpenURL } from '../../../wailsjs/runtime/runtime';
import BlueSkyLogo from '../../assets/icons/bluesky.svg';
import DiscordIcon from '../../assets/icons/discord.svg';
import GithubIcon from '../../assets/icons/github.svg';
import OpenCollectiveIcon from '../../assets/icons/opencollective.svg';
import RedditIcon from '../../assets/icons/reddit.svg';
import { SOCIAL_LINKS } from '../../constants/urls';

export function ConnectScreen() {
  const { t } = useTranslation();

  return (
    <div className="intro-screen">
      <h2 className="intro-heading">{t('introOverlay.screen4.title')}</h2>
      <p>{t('introOverlay.screen4.description')}</p>

      <Card elevation={1} className="connect-card">
        <H5>{t('introOverlay.screen4.socialText')}</H5>

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

        <p>{t('introOverlay.screen4.donateText')}</p>
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
